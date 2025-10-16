package git

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func commit(ctx context.Context) error {
	out, err := utils.ExecCommand(ctx, "git", "diff", "--staged", "--shortstat")
	if err != nil {
		return errors.Wrap(err, "git diff")
	}

	if out == "" {
		return nil
	}

	_, err = utils.ExecCommand(ctx, "git", "commit", "-m", "undefined")
	if err != nil {
		_, err = utils.ExecCommand(ctx, "git", "commit", "-m", "undefined")
		if err != nil {
			return errors.Wrap(err, "git commit")
		}
	}

	return nil
}

func RunGitAuto(ctx context.Context, _ *cli.Command) error {
	_, err := utils.ExecCommand(ctx, "git", "add", ".")
	if err != nil {
		return errors.Wrap(err, "git add .")
	}

	err = commit(ctx)
	if err != nil {
		return errors.Wrap(err, "git add .")
	}

	_, err = utils.ExecCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "git push")
	}
	// TODO don't raise error on no push

	return nil
}
