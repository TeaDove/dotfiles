package cli

import (
	"context"
	"dotfiles/pkg/http_supplier"
	"fmt"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func (r *CLI) commandLocateByIP(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() == 0 {
		return errors.New("need at least one IP address or domain")
	}

	ipOrDomain := command.Args().First()

	domainLocation, err := http_supplier.New().LocateByIP(ctx, ipOrDomain)
	if err != nil {
		return errors.Wrap(err, "failed to get resp")
	}

	fmt.Printf( //nolint:forbidigo // is ok
		"Address: %s %s %s\n",
		color.GreenString(domainLocation.Country),
		domainLocation.RegionName,
		domainLocation.City,
	)
	fmt.Printf( //nolint:forbidigo // is ok
		"Coordinates: %s (https://yandex.ru/maps/?ll=%f%%2C%f&z=16)\n",
		color.YellowString(fmt.Sprintf("%f,%f", domainLocation.Lat, domainLocation.Lon)),
		domainLocation.Lon,
		domainLocation.Lat,
	)
	fmt.Printf( //nolint:forbidigo // is ok
		"Organization: %s, %s, %s\n",
		domainLocation.Isp,
		color.BlueString(domainLocation.Org),
		domainLocation.As,
	)

	return nil
}
