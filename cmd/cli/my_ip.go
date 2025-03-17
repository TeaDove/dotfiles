package cli

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func (r *CLI) commandMyIP(ctx context.Context, command *cli.Command) error {
	ip, err := r.httpSupplier.MyIP(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch ip")
	}

	fmt.Print(ip)
	return nil
}
