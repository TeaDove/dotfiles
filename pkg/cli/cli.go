package cli

import (
	"context"
	"dotfiles/pkg/cli/net_stats"
	"dotfiles/pkg/cli/watch"
	"dotfiles/pkg/http_supplier"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

type CLI struct{}

func NewCLI() *CLI {
	return &CLI{}
}

var verboseFlag = &cli.BoolFlag{Name: "v", Usage: "verbose info"} //nolint:gochecknoglobals // is ok

func (r *CLI) Run(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		return errors.New("go fuck yourself with windows OS")
	}

	cmd := &cli.Command{
		Name:        "dotfiles",
		Description: "set of useful command",
		Flags:       []cli.Flag{verboseFlag},
		Commands: []*cli.Command{
			{
				Name:   "install",
				Usage:  "install all dotfiles, i.e. fish config",
				Action: r.commandInstall,
			},
			{
				Name:   "update",
				Usage:  "updates this executable",
				Action: r.commandUpdate,
			},
			{
				Name:   "u",
				Usage:  "generates random uuid",
				Action: r.commandUUID,
			},
			{
				Name:   "t",
				Usage:  "generates save to use password",
				Action: r.commandText,
			},
			{
				Name:   "l",
				Usage:  "locates service by ip or domain from http://ip-api.com/json/",
				Action: r.commandLocateByIP,
			},
			{
				Name:   "net",
				Usage:  "displays all network stats",
				Action: r.commandNet,
			},
			{
				Name:   "sha",
				Usage:  "hashes string as sha512",
				Action: r.commandSha,
			},
			{
				Name:   "git-pull-and-merge",
				Usage:  "git utils",
				Action: r.commandGitPullAndMerge,
			},
			{
				Name:   "watch",
				Usage:  "like unix watch, but better",
				Action: r.commandWatch,
				Flags:  []cli.Flag{watch.IntervalFlag},
			},
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		return errors.Wrap(err, "failed to run cmd")
	}

	return nil
}

func (r *CLI) commandNet(ctx context.Context, _ *cli.Command) error {
	return net_stats.NewNetStats(http_supplier.New()).Run(ctx)
}

func (r *CLI) commandWatch(ctx context.Context, cmd *cli.Command) error {
	return watch.New().Run(ctx, cmd)
}
