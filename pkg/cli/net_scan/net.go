package net_scan

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"fmt"
	netstd "net"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/urfave/cli/v3"
)

type NetSystem struct{}

func Run(ctx context.Context, _ *cli.Command) error {
	mainInterface, err := getMainInterface(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	_, ipnet, err := netstd.ParseCIDR(mainInterface.Addrs[0].Addr)
	if err != nil {
		return errors.Wrap(err, "parse main address")
	}

	fmt.Printf("Checking %s interface on %s network\n\n", color.YellowString(mainInterface.Name), ipnet.String())

	var (
		wg        sync.WaitGroup
		semaphore = utils.NewSemaphore(1000)
	)
	for ip := range iterateOverNet(ipnet) {
		wg.Go(func() {
			checkAddress(ctx, semaphore, ip)
		})
	}

	wg.Wait()

	return nil
}

func getMainInterface(ctx context.Context) (net.InterfaceStat, error) {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return net.InterfaceStat{}, errors.Wrap(err, "get interfaces")
	}

	if len(interfaces) <= 1 {
		return net.InterfaceStat{}, errors.New("only loopback found")
	}

	var mainInterface net.InterfaceStat

	for _, i := range interfaces[1:] {
		if len(i.Addrs) != 0 {
			mainInterface = i
			break
		}
	}

	if mainInterface.Name == "" {
		return net.InterfaceStat{}, errors.New("no interfaces with addresses found")
	}

	return mainInterface, nil
}

func checkAddress(ctx context.Context, semaphore *utils.Semaphore, ip netstd.IP) {
	var ok bool

	semaphore.Locked(func() {
		ok = ping(ctx, ip.String())
	})

	if !ok {
		return
	}

	fmt.Printf("%s is pingable!\n", color.GreenString(ip.String()))

	var wg sync.WaitGroup

	for i := 1; i < 10000; i++ {
		wg.Go(func() {
			semaphore.Locked(func() {
				dialer := netstd.Dialer{Timeout: 5 * time.Second}

				conn, err := dialer.DialContext(ctx, "tcp", ip.String()+":"+strconv.Itoa(i))
				if err != nil {
					return
				}

				defer conn.Close()

				fmt.Printf("%s:%d is open!\n", color.GreenString(ip.String()), i)
			})
		})
	}

	wg.Wait()
}
