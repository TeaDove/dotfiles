package cli

import (
	"context"
	"dotfiles/pkg/cli/kwatch"
	"dotfiles/pkg/cli/net_stats"
	"dotfiles/pkg/http_supplier"
	"dotfiles/pkg/kube_supplier"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

type CLI struct {
	httpSupplier *http_supplier.Supplier
}

func NewCLI() *CLI {
	return &CLI{httpSupplier: http_supplier.New()}
}

var verboseFlag = &cli.BoolFlag{Name: "v", Usage: "verbose info"}

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
				Usage:  "Displays all network stats",
				Action: r.commandNet,
			},
			{
				Name:   "sha",
				Usage:  "Hashes string as sha512",
				Action: r.commandSha,
			},
			{
				Name:   "jq",
				Usage:  "Attempt to parse every line as json and write it",
				Action: r.commandJq,
			},
			{
				Name:   "git-pull-and-merge",
				Usage:  "Git utils",
				Action: r.commandGitPullAndMerge,
			},
			{
				Name:   "kwatch",
				Usage:  "Displays k8s pod usage in current namespace",
				Action: r.commandKwatch,
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
	return net_stats.NewNetStats(r.httpSupplier).Run(ctx)
}

func (r *CLI) commandKwatch(ctx context.Context, _ *cli.Command) error {
	kubeSupplier, err := kube_supplier.NewSupplier()
	if err != nil {
		return errors.Wrap(err, "failed to init kube_supplier")
	}

	return kwatch.New(kubeSupplier).Run(ctx)
}
