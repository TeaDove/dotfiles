package git

import (
	"context"
	"dotfiles/pkg/cli/utils"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

var NoVerifyFlag = &cli.BoolFlag{Name: "no-verify", Usage: "no verify"} //nolint:gochecknoglobals // is ok

func RunGitAuto(ctx context.Context, cmd *cli.Command) error {
	_, err := utils.ExecCommand(ctx, "git", "add", ".")
	if err != nil {
		return errors.Wrap(err, "git add")
	}

	noVerify := cmd.Bool(NoVerifyFlag.Name)

	err = commit(ctx, noVerify)
	if err != nil {
		return errors.Wrap(err, "git commit")
	}

	// TODO don't raise error on no push
	_, err = utils.ExecCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "git push")
	}

	return nil
}

func commit(ctx context.Context, noVerify bool) error {
	out, err := utils.ExecCommand(ctx, "git", "diff", "--staged", "--shortstat")
	if err != nil {
		return errors.Wrap(err, "git diff")
	}

	if out == "" {
		return nil
	}

	args := []string{"commit", "-m", getCommitMsg(ctx)}
	if noVerify {
		args = append(args, "--no-verify")
	}

	_, err = utils.ExecCommand(ctx, "git", args...)
	if err != nil {
		_, err = utils.ExecCommand(ctx, "git", args...)
		if err != nil {
			return errors.Wrap(err, "git commit")
		}
	}

	return nil
}
