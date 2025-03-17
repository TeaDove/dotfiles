package cli

import (
	"context"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"os"
	"runtime"
)

type CLI struct{}

func NewCLI() *CLI {
	return &CLI{}
}

func (r *CLI) Run(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		return errors.New("go fuck yourself with windows OS")
	}

	cmd := &cli.Command{
		Name:        "dotfiles",
		Description: "set of useful command",
		Commands: []*cli.Command{
			{
				Name:        "install",
				Description: "install all dotfiles like fish config",
				Action:      r.commandInstall,
			},
			{
				Name:        "u",
				Description: "generates random uuid",
				Action:      r.commandUuid,
			},
			{
				Name:        "l",
				Description: "locates service by ip or domain from http://ip-api.com/json/",
				Action:      r.commandLocateByIP,
			},
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		return errors.Wrap(err, "failed to run cmd")
	}

	return nil
}
