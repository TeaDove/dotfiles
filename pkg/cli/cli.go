package cli

import (
	"context"
	net_sniff "dotfiles/pkg/cli/net_scan"
	net_scan "dotfiles/pkg/cli/net_sniff"
	"dotfiles/pkg/cli/net_system"
	"dotfiles/pkg/cli/watch"
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
				Name:  "net",
				Usage: "net utils",
				Commands: []*cli.Command{
					{
						Name:   "system",
						Usage:  "display information about this machine",
						Action: net_system.Run,
					},
					{
						Name:   "scan",
						Usage:  "display information about local networks",
						Action: net_scan.Run,
					},
					{
						Name:   "sniff",
						Usage:  "sniff traffic!",
						Action: net_sniff.Run,
					},
				},
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
				Action: watch.Run,
				Flags:  []cli.Flag{watch.IntervalFlag},
			},
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		return errors.Wrap(err, "run cmd")
	}

	return nil
}
