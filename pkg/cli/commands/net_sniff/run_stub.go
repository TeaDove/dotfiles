//go:build darwin || !darwin

package net_sniff

import (
	"context"

	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, _ *cli.Command) error {
	panic("Implemented only for darwin")
}
