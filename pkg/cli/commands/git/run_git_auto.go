package git

import (
	"context"
	"dotfiles/pkg/cli/utils"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func RunGitAuto(ctx context.Context, _ *cli.Command) error {
	_, err := utils.ExecCommand(ctx, "git", "add", ".")
	if err != nil {
		return errors.Wrap(err, "git add .")
	}

	_, err = utils.ExecCommand(ctx, "git", "commit", "-m", "undefined")
	if err != nil {
		_, _ = utils.ExecCommand(ctx, "git", "commit", "-m", "undefined")
	}

	_, err = utils.ExecCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "git push")
	}
	// TODO don't raise error on no push

	return nil
}
