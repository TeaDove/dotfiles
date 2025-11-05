package cli

import (
	"context"
	"dotfiles/pkg/cli/commands/git"
	"dotfiles/pkg/cli/commands/net_scan"
	"dotfiles/pkg/cli/commands/net_sniff"
	"dotfiles/pkg/cli/commands/net_system"
	"dotfiles/pkg/cli/commands/watch"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

var verboseFlag = &cli.BoolFlag{Name: "v", Usage: "verbose info"} //nolint:gochecknoglobals // is ok

func Run(ctx context.Context) error {
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
				Action: CommandInstall,
			},
			{
				Name:   "update",
				Usage:  "updates this executable",
				Action: CommandUpdate,
			},
			{
				Name:   "u",
				Usage:  "generates random uuid",
				Action: CommandUUID,
			},
			{
				Name:   "t",
				Usage:  "generates save to use password",
				Action: CommandText,
			},
			{
				Name:   "l",
				Usage:  "locates service by ip or domain from http://ip-api.com/json/",
				Action: CommandLocateByIP,
			},
			{
				Name:  "g",
				Usage: "git utils",
				Commands: []*cli.Command{
					{
						Name:   "a",
						Usage:  "add, commit and push",
						Action: git.RunGitAuto,
						Flags:  []cli.Flag{git.NoVerifyFlag},
					},
					{
						Name:   "m",
						Usage:  "merge from master",
						Action: git.RunGitPullAndMerge,
					},
				},
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
				Action: CommandSha,
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
