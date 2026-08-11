package git

import (
	"context"
	"dotfiles/pkg/cli/utils/systemutils"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v3"
)

func RunGitPullAndMerge(ctx context.Context, _ *cli.Command) error {
	const master = "master"

	out, err := systemutils.ExecCommand(ctx, "git", "status", "-s")
	if err != nil {
		return errors.Wrap(err, "get current branch")
	}

	if strings.TrimSpace(out) != "" {
		return errors.New("you have unpushed changes")
	}

	out, err = systemutils.ExecCommand(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return errors.Wrap(err, "get current branch")
	}

	branch := strings.TrimSpace(out)
	if branch == master {
		return errors.New("current branch is master")
	}

	_, err = systemutils.ExecCommand(ctx, "git", "checkout", master)
	if err != nil {
		return errors.Wrap(err, "checkout master")
	}

	_, err = systemutils.ExecCommand(ctx, "git", "pull")
	if err != nil {
		return errors.Wrap(err, "pull")
	}

	_, err = systemutils.ExecCommand(ctx, "git", "checkout", branch)
	if err != nil {
		return errors.Wrap(err, "checkout")
	}

	_, err = systemutils.ExecCommand(ctx, "git", "merge", master)
	if err != nil {
		return errors.Wrap(err, "merge")
	}

	_, err = systemutils.ExecCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "push")
	}

	return nil
}
