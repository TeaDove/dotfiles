package cli

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"strings"
)

func (r *CLI) commandGitPullAndMerge(ctx context.Context, _ *cli.Command) error {
	const master = "master"

	out, err := utils.ExecCommand(ctx, "git", "status", "-s")
	if err != nil {
		return errors.Wrap(err, "failed to get current branch")
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("you have unpushed changes")
	}

	out, err = utils.ExecCommand(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return errors.Wrap(err, "failed to get current branch")
	}
	branch := strings.TrimSpace(out)
	if branch == master {
		return errors.New("current branch is master")
	}

	_, err = utils.ExecCommand(ctx, "git", "checkout", master)
	if err != nil {
		return errors.Wrap(err, "failed to checkout master")
	}

	_, err = utils.ExecCommand(ctx, "git", "pull")
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	_, err = utils.ExecCommand(ctx, "git", "checkout", branch)
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	out, err = utils.ExecCommand(ctx, "git", "merge", master)
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	out, err = utils.ExecCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	return nil
}
