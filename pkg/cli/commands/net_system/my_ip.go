package net_system

import (
	"context"
	"dotfiles/pkg/http_supplier"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
)

func (r *Service) myIPView(ctx context.Context) string {
	myIP, err := r.httpSupplier.MyIP(ctx)
	if err != nil {
		return prettyErr(errors.Wrap(err, "get my-ip"))
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s: %s", color.GreenString("My IP"), color.YellowString(myIP.String())))

	builder.WriteString(fmt.Sprintf(" (%s)\n", r.shortLocationOrErr(ctx, myIP.String())))

	builder.WriteString(color.GreenString("DNS Servers: "))

	dnss := http_supplier.GetDNSServers()

	var dnssStrings []string

	for _, dns := range dnss {
		if !dns.Addr().IsPrivate() && !dns.Addr().IsLoopback() {
			dnssStrings = append(
				dnssStrings,
				fmt.Sprintf("%s (%s)", dns.String(), r.shortLocationOrErr(ctx, dns.Addr().String())),
			)

			continue
		}

		dnssStrings = append(dnssStrings, dns.String())
	}

	if len(dnssStrings) == 1 {
		builder.WriteString(dnssStrings[0])
	} else {
		builder.WriteString("\n")
		builder.WriteString(strings.Join(dnssStrings, "\n"))
	}

	return builder.String()
}

func (r *Service) shortLocationOrErr(ctx context.Context, ipOrDomain string) string {
	location, err := r.httpSupplier.LocateByIP(ctx, ipOrDomain)
	if err != nil {
		return prettyErr(errors.Wrap(err, "get location"))
	}

	return fmt.Sprintf("%s, %s, %s", location.Country, location.City, color.BlueString(location.Org))
}
