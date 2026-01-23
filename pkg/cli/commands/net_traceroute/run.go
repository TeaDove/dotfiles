package net_traceroute

import (
	"context"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/teadove/teasutils/utils/time_utils"

	"github.com/urfave/cli/v3"

	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func Run(ctx context.Context, command *cli.Command) error {
	const (
		maxHops  = 128
		timeout  = 2 * time.Second
		basePort = 33434
	)

	target := command.Args().First()

	dstIP := net.ParseIP(target).To4()
	if dstIP == nil {
		return errors.Newf("invalid destination IP %q", target)
	}

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

	fmt.Printf("traceroute to %s (%s), %d hops max\n", target, dstIP.String(), maxHops)
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

		var builder strings.Builder
		builder.WriteString(strconv.Itoa(ttl))

		if result.peer != nil {
			builder.WriteString(fmt.Sprintf(" %s %s", result.peer.String(), time_utils.RoundDuration(result.rtt)))

			if dstIP.String() == result.peer.String() {
				fmt.Println(builder.String())
				return nil
			}
		}

		if result.icmpErr != "" {
			builder.WriteString(fmt.Sprintf(" %s", result.icmpErr))
		}

		fmt.Println(builder.String())
	}

	fmt.Println("max hops reached")

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
	peer    net.Addr
	rtt     time.Duration
	icmpErr string
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
		peer:    peer,
		rtt:     time.Since(start),
		icmpErr: "",
	}

	if err != nil {
		result.icmpErr = "* * *"
		return result, nil //nolint: nilerr // as expected
	}

	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), buf[:n])
	if err != nil {
		result.icmpErr = fmt.Sprintf("parse error: %s", err.Error())
		return result, nil
	}

	switch msg.Type {
	case ipv4.ICMPTypeTimeExceeded:
		return result, nil
	case ipv4.ICMPTypeDestinationUnreachable:
		if msg.Code == 3 && peer.String() == r.dstIP.String() {
			return result, nil
		}

		result.icmpErr = strconv.Itoa(msg.Code)

		return result, nil

	default:
		result.icmpErr = fmt.Sprintf("%v", msg.Type)
		return result, nil
	}
}
