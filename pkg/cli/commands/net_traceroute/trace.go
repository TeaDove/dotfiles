package net_traceroute

import (
	"context"
	"dotfiles/pkg/http_supplier"
	"fmt"
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
		result, err := r.trace(ctx, uint16(ttl))
		if err != nil {
			return errors.Wrap(err, "trace")
		}

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
	ttl     uint16
	peer    net.Addr
	rtt     time.Duration
	icmpErr string

	domains  Result[string]
	location Result[http_supplier.DomainLocationResp]
}

type Result[T any] struct {
	Ok  T
	Err error
}

func (r *Service) trace(ctx context.Context, ttl uint16) (traceResult, error) {
	udpConn, err := (&net.ListenConfig{}).ListenPacket(ctx, "udp4", "0.0.0.0:0")
	if err != nil {
		return traceResult{}, errors.Wrap(err, "udp4 listen packet")
	}
	defer udpConn.Close()

	v4udp := ipv4.NewPacketConn(udpConn)
	defer v4udp.Close()

	if err = v4udp.SetTTL(int(ttl)); err != nil {
		return traceResult{}, errors.Wrap(err, "udp set ttl")
	}

	port := r.basePort + ttl
	dst := &net.UDPAddr{IP: r.dstIP, Port: int(port)}

	start := time.Now()

	_, err = udpConn.WriteTo([]byte{0x42}, dst)
	if err != nil {
		return traceResult{}, errors.Wrap(err, "udp conn write")
	}

	buf := make([]byte, 1500)

	err = r.icmpConn.SetReadDeadline(time.Now().Add(r.timeout))
	if err != nil {
		return traceResult{}, errors.Wrap(err, "icmp set read deadline")
	}

	n, peer, err := r.icmpConn.ReadFrom(buf)
	result := traceResult{
		ttl:  ttl,
		peer: peer,
		rtt:  time.Since(start),
	}

	if err != nil {
		return result, nil
	}

	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), buf[:n])
	if err != nil {
		result.icmpErr = fmt.Sprintf("parse error: %s", err.Error())
		return result, nil
	}

	switch msg.Type {
	case ipv4.ICMPTypeTimeExceeded:
		// Expected hop
		return result, nil
	case ipv4.ICMPTypeDestinationUnreachable:
		if msg.Code == 3 && peer.String() == r.dstIP.String() {
			// reached
			return result, nil
		}

		result.peer = nil
		result.icmpErr = fmt.Sprintf("msg.Code=%d", msg.Code)

		return result, nil

	default:
		result.peer = nil
		result.icmpErr = fmt.Sprintf("msg.Type=%v", msg.Type)

		return result, nil
	}
}

func getUDPPort(msg *icmp.Message) (uint16, error) {
	var data []byte

	switch b := msg.Body.(type) {
	case *icmp.TimeExceeded:
		data = b.Data
	case *icmp.DstUnreach:
		data = b.Data
	default:
		return 0, errors.New("unexpected message type")
	}

	iphdr, err := ipv4.ParseHeader(data)
	if err != nil {
		return 0, errors.Wrap(err, "parse header")
	}

	if iphdr.Protocol != 17 {
		return 0, errors.Newf("wrong protocol %d", iphdr.Protocol)
	}

	off := iphdr.Len
	if len(data) < off+8 {
		return 0, errors.Newf("data too short")
	}

	dstPort := int(data[off+2])<<8 | int(data[off+3])

	return uint16(dstPort), nil
}

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
