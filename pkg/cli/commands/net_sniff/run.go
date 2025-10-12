package net_sniff

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, _ *cli.Command) error {
	panic("Not implemented")

	devices, err := pcap.FindAllDevs()
	if err != nil {
		return errors.WithStack(err)
	}

	if len(devices) == 0 {
		return errors.New("no devices found")
	}

	device := devices[0]
	fmt.Printf("using %s device %v\n", color.CyanString(device.Name), device.Addresses)

	// Open the device for capturing
	handle, err := pcap.OpenLive(device.Name, 1600, true, 30*time.Second)
	if err != nil {
		return errors.WithStack(err)
	}
	defer handle.Close()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		ipv4, ok := packet.NetworkLayer().(*layers.IPv4)
		if !ok {
			continue
		}

		tcp, ok := packet.TransportLayer().(*layers.TCP)
		if !ok {
			continue
		}

		if tcp.SrcPort != 80 && tcp.DstPort != 80 {
			continue
		}

		fmt.Printf(
			"%s:%d -> %s:%d\n",
			color.CyanString(ipv4.SrcIP.String()),
			tcp.SrcPort,
			color.CyanString(ipv4.DstIP.String()),
			tcp.DstPort,
		)
		println(string(tcp.Payload))
		// fmt.Printf("%s -> %s\n", color.CyanString(ipv4.SrcIP.String()), color.CyanString(ipv4.DstIP.String()))
	}

	return nil
}
