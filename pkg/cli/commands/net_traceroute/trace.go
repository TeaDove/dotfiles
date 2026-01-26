package net_traceroute

import (
	"context"
	"dotfiles/pkg/http_supplier"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func parseAddr(ctx context.Context, target string) (net.IP, error) {
	dstIP := net.ParseIP(target).To4()
	if dstIP == nil {
		addrs, err := (&net.Resolver{}).LookupHost(ctx, target)
		if err != nil {
			return net.IP{}, errors.Wrap(err, "bad ip or domain")
		}

		if len(addrs) == 0 {
			return net.IP{}, errors.New("no ip found")
		}

		dstIP = net.ParseIP(addrs[0]).To4()
		if dstIP == nil {
			return net.IP{}, errors.Newf("invalid destination IP %q", target)
		}
	}

	return dstIP, nil
}

func (r *Service) traceRoute(ctx context.Context, dstIP net.IP) error {
	icmpConn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return errors.Wrap(err, "listen for icmp packets, most likely you need to run it with sudo: sudo !!")
	}
	defer icmpConn.Close()

	r.icmpConn = icmpConn

	for ttl := 1; ttl <= r.maxHops; ttl++ {
		start, err := r.sendTrace(ctx, uint16(ttl))
		if err != nil {
			return errors.Wrap(err, "send trace")
		}

		result := r.catchTrace(start, uint16(ttl))
		r.populateTraceResult(ctx, &result)
		r.hopsMu.Lock()
		r.hops = append(r.hops, result)
		slices.SortFunc(r.hops, func(a, b traceResult) int {
			if a.ttl < b.ttl {
				return -1
			}

			if a.ttl > b.ttl {
				return 1
			}

			return 0
		})
		r.model.updateTable()
		r.hopsMu.Unlock()

		if result.peer != nil && dstIP.String() == result.peer.String() {
			return nil
		}
	}

	return nil
}

type traceResult struct {
	ttl  uint16
	peer net.Addr
	rtt  time.Duration
	err  error

	domains  Result[string]
	location Result[http_supplier.DomainLocationResp]
}

type Result[T any] struct {
	Ok  T
	Err error
}

func (r *Service) sendTrace(ctx context.Context, ttl uint16) (time.Time, error) {
	udpConn, err := (&net.ListenConfig{}).ListenPacket(ctx, "udp4", "0.0.0.0:0")
	if err != nil {
		return time.Time{}, errors.Wrap(err, "udp4 listen packet")
	}
	defer udpConn.Close()

	v4udp := ipv4.NewPacketConn(udpConn)
	defer v4udp.Close()

	if err = v4udp.SetTTL(int(ttl)); err != nil {
		return time.Time{}, errors.Wrap(err, "udp set ttl")
	}

	port := r.basePort + ttl
	dst := &net.UDPAddr{IP: r.dstIP, Port: int(port)}

	start := time.Now()

	_, err = udpConn.WriteTo([]byte{0x42}, dst)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "udp conn write")
	}

	return start, nil
}

func (r *Service) catchTrace(start time.Time, ttl uint16) traceResult {
	result := traceResult{ttl: ttl}
	buf := make([]byte, 1500)

	err := r.icmpConn.SetReadDeadline(time.Now().Add(r.timeout))
	if err != nil {
		result.err = errors.Wrap(err, "icmp set read deadline")
		return result
	}

	n, peer, err := r.icmpConn.ReadFrom(buf)
	result.rtt = time.Since(start)

	if err != nil {
		result.err = errors.Wrap(err, "icmp read")
		return result
	}

	result.peer = peer

	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), buf[:n])
	if err != nil {
		result.err = errors.Wrap(err, "parse error")
		return result
	}

	switch msg.Type {
	case ipv4.ICMPTypeTimeExceeded:
		// Expected hop
		return result
	case ipv4.ICMPTypeDestinationUnreachable:
		if msg.Code == 3 && peer.String() == r.dstIP.String() {
			// reached
			return result
		}

		result.peer = nil
		result.err = errors.Newf("msg.Code=%d", msg.Code)

		return result

	default:
		result.peer = nil
		result.err = errors.Newf("msg.Type=%v", msg.Type)

		return result
	}
}

// func (r *Service) trace(ctx context.Context, ttl uint16) (traceResult, error) {
//	start, err := r.sendTrace(ctx, ttl)
//	if err != nil {
//		return traceResult{}, errors.Wrap(err, "send trace")
//	}
//
//	buf := make([]byte, 1500)
//
//	err = r.icmpConn.SetReadDeadline(time.Now().Add(r.timeout))
//	if err != nil {
//		return traceResult{}, errors.Wrap(err, "icmp set read deadline")
//	}
//
//	n, peer, err := r.icmpConn.ReadFrom(buf)
//	result := traceResult{
//		ttl:  ttl,
//		peer: peer,
//		rtt:  time.Since(start),
//	}
//
//	if err != nil {
//		return result, nil
//	}
//
//	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), buf[:n])
//	if err != nil {
//		result.icmpErr = fmt.Sprintf("parse error: %s", err.Error())
//		return result, nil
//	}
//
//	udpPort, err := getUDPPort(msg)
//	if err != nil {
//		result.icmpErr = fmt.Sprintf("get UDP error: %s", err.Error())
//		return result, nil
//	}
//
//	result.ttl = udpPort - r.basePort
//
//	switch msg.Type {
//	case ipv4.ICMPTypeTimeExceeded:
//		// Expected hop
//		return result, nil
//	case ipv4.ICMPTypeDestinationUnreachable:
//		if msg.Code == 3 && peer.String() == r.dstIP.String() {
//			// reached
//			return result, nil
//		}
//
//		result.peer = nil
//		result.icmpErr = fmt.Sprintf("msg.Code=%d", msg.Code)
//
//		return result, nil
//
//	default:
//		result.peer = nil
//		result.icmpErr = fmt.Sprintf("msg.Type=%v", msg.Type)
//
//		return result, nil
//	}
//}

func (r *Service) populateTraceResult(ctx context.Context, result *traceResult) {
	if result.peer == nil {
		return
	}

	ip := result.peer.String()

	addresses, err := (&net.Resolver{}).LookupAddr(ctx, ip)
	if err != nil {
		result.domains.Err = err
	} else {
		result.domains.Ok = strings.Join(addresses, ",")
	}

	location, err := r.httpSupplier.LocateByIP(ctx, ip)
	if err != nil {
		result.location.Err = err
	} else {
		result.location.Ok = location
	}
}
