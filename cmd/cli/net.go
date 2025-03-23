package cli

import (
	"context"
	"dotfiles/cmd/cli/net_stats"
	"github.com/urfave/cli/v3"
)

func (r *CLI) commandNet(ctx context.Context, command *cli.Command) error {
	return net_stats.NewNetStats(r.httpSupplier).Net(ctx)
}
