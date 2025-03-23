package net_stats

import (
	"context"
	"dotfiles/cmd/http_supplier"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"sync"
)

func (r *NetStats) myIPView(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	myIp, err := r.httpSupplier.MyIP(ctx)
	if err != nil {
		r.model.myIP = prettyErr(errors.Wrap(err, "failed to get my-ip"))
		return
	}

	r.model.myIP = ""
	r.model.myIP += fmt.Sprintf("%s: %s", color.GreenString("My IP"), color.YellowString(myIp.String()))

	r.model.myIP += fmt.Sprintf(" (%s)", r.shortLocationOrErr(ctx, myIp.String()))

	r.model.myIP += "\n"
	r.model.myIP += color.GreenString("DNS Servers: ")

	dnss := http_supplier.GetDNSServers()

	for _, dns := range dnss {
		if !dns.Addr().IsGlobalUnicast() {
			r.model.myIP += fmt.Sprintf("%s (%s)", dns.String(), r.shortLocationOrErr(ctx, dns.Addr().String()))
			continue
		}

		r.model.myIP += dns.String()
	}
}
