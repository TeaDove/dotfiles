package git

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func RunGitAuto(ctx context.Context, _ *cli.Command) error {
	_, err := utils.ExecCommand(ctx, "git", "add", ".")
	if err != nil {
		return errors.Wrap(err, "git add .")
	}

	_, err = utils.ExecCommand(ctx, "git", "commit", "-m", `"auto commit"`)
	if err != nil {
		return errors.Wrap(err, "git commit")
	}

	_, err = utils.ExecCommand(ctx, "git", "commit", "-m", "undefined")
	if err != nil {
		var exitErr exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return errors.Wrap(err, "git commit")
		}

		_, err = utils.ExecCommand(ctx, "git", "commit", "-m", "undefined")
		if err != nil {
			return errors.Wrap(err, "git commit")
		}
	}

	_, err = utils.ExecCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "git push")
	}

	return nil
}
