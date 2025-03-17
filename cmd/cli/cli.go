package cli

import (
	"context"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"os"
)

type CLI struct{}

func NewCLI() *CLI {
	return &CLI{}
}

func (r *CLI) Run(ctx context.Context) error {
	cmd := &cli.Command{
		Name:        "dotfiles",
		Description: "set of useful command",
		Commands: []*cli.Command{
			{
				Name:        "install",
				Description: "install all dotfiles like fish config",
				Action:      r.commandInstall,
			},
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		return errors.Wrap(err, "failed to run cmd")
	}

	return nil
}
