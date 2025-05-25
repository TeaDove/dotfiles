package net_stats

import (
	"context"
	"dotfiles/pkg/http_supplier"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

func (r *NetStats) myIPView(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	myIP, err := r.httpSupplier.MyIP(ctx)
	if err != nil {
		r.model.myIP = prettyErr(errors.Wrap(err, "failed to get my-ip"))
		return
	}

	r.model.myIP = ""
	r.model.myIP += fmt.Sprintf("%s: %s", color.GreenString("My IP"), color.YellowString(myIP.String()))

	r.model.myIP += fmt.Sprintf(" (%s)", r.shortLocationOrErr(ctx, myIP.String()))

	r.model.myIP += "\n"
	r.model.myIP += color.GreenString("DNS Servers: ")

	dnss := http_supplier.GetDNSServers()
	var dnssStrings []string
	for _, dns := range dnss {
		if !dns.Addr().IsPrivate() && !dns.Addr().IsLoopback() {
			dnssStrings = append(dnssStrings, fmt.Sprintf("%s (%s)", dns.String(), r.shortLocationOrErr(ctx, dns.Addr().String())))
			continue
		}

		dnssStrings = append(dnssStrings, dns.String())
	}

	if len(dnssStrings) == 1 {
		r.model.myIP += dnssStrings[0]
	} else {
		r.model.myIP += "\n"
		r.model.myIP += strings.Join(dnssStrings, "\n")
	}

}

func (r *NetStats) shortLocationOrErr(ctx context.Context, ipOrDomain string) string {
	location, err := r.httpSupplier.LocateByIP(ctx, ipOrDomain)
	if err != nil {
		return prettyErr(errors.Wrap(err, "failed to get location"))
	}

	return fmt.Sprintf("%s, %s, %s", location.Country, location.City, color.BlueString(location.Org))
}
