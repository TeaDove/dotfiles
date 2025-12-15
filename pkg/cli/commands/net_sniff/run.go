//go:build darwin && !darwin

package net_sniff

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func getMainDevice() (pcap.Interface, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return pcap.Interface{}, errors.WithStack(err)
	}

	for _, device := range devices {
		for _, address := range device.Addresses {
			if len(address.Broadaddr) != 0 {
				return device, nil
			}
		}
	}

	return pcap.Interface{}, errors.New("no devices found")
}

func Run(ctx context.Context, _ *cli.Command) error {
	device, err := getMainDevice()
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("using %s device %v\n", color.CyanString(device.Name), device.Addresses)

	// Open the device for capturing
	handle, err := pcap.OpenLive(device.Name, 1600, true, pcap.BlockForever)
	if err != nil {
		return errors.WithStack(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("tcp")
	if err != nil {
		return errors.WithStack(err)
	}

	// Use the handle as a packet source to process all packets

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		for _, layer := range packet.Layers() {
			fmt.Printf("%s ", layer.LayerType().String())
		}
		fmt.Println()

		ipv4, ok := packet.NetworkLayer().(*layers.IPv4)
		if !ok {
			continue
		}

		tcp, ok := packet.TransportLayer().(*layers.TCP)
		if !ok {
			continue
		}

		fmt.Printf(
			"%s:%d -> %s:%d\n",
			color.CyanString(ipv4.SrcIP.String()),
			tcp.SrcPort,
			color.CyanString(ipv4.DstIP.String()),
			tcp.DstPort,
		)
		// fmt.Printf("%s -> %s\n", color.CyanString(ipv4.SrcIP.String()), color.CyanString(ipv4.DstIP.String()))
	}

	return nil
}
