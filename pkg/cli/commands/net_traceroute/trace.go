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
	const (
		maxHops  = 128
		timeout  = 1 * time.Second
		basePort = 33434
	)

	icmpConn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return errors.Wrap(err, "listen for icmp packets, most likely you need to run it with sudo: sudo !!")
	}
	defer icmpConn.Close()

	udpConn, err := (&net.ListenConfig{}).ListenPacket(ctx, "udp4", "0.0.0.0:0")
	if err != nil {
		return errors.Wrap(err, "udp4 listen packet")
	}
	defer udpConn.Close()

	v4udp := ipv4.NewPacketConn(udpConn)

	conn := Conn{
		v4udp:    v4udp,
		udpConn:  udpConn,
		icmpConn: icmpConn,
		basePort: basePort,
		dstIP:    dstIP,
		timeout:  timeout,
		ttl:      0,
	}

	for ttl := 1; ttl <= maxHops; ttl++ {
		conn.ttl = ttl

		result, err := conn.trace()
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

type Conn struct {
	v4udp    *ipv4.PacketConn
	udpConn  net.PacketConn
	icmpConn *icmp.PacketConn

	basePort int
	dstIP    net.IP
	timeout  time.Duration

	ttl int
}

type traceResult struct {
	ttl     int
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

func (r *Conn) trace() (traceResult, error) {
	if err := r.v4udp.SetTTL(r.ttl); err != nil {
		return traceResult{}, errors.Wrap(err, "udp set ttl")
	}

	dst := &net.UDPAddr{IP: r.dstIP, Port: r.basePort + r.ttl}

	start := time.Now()

	_, err := r.udpConn.WriteTo([]byte{0x42}, dst)
	if err != nil {
		return traceResult{}, errors.Wrap(err, "udp conn write")
	}

	buf := make([]byte, 1500)

	_ = r.icmpConn.SetReadDeadline(time.Now().Add(r.timeout))

	n, peer, err := r.icmpConn.ReadFrom(buf)
	result := traceResult{
		ttl:  r.ttl,
		peer: peer,
		rtt:  time.Since(start),
	}

	if err != nil {
		return result, nil //nolint: nilerr // as expected
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
