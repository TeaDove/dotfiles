package cli

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"os/exec"
	"strings"
)

func execCommand(ctx context.Context, name string, args ...string) (string, error) {
	color.Magenta(fmt.Sprintf("$ %s %s", name, strings.Join(args, " ")))
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current branch")
	}

	color.White(string(out))

	return string(out), nil
}

func (r *CLI) commandGitPullAndMerge(ctx context.Context, command *cli.Command) error {
	const master = "master"

	out, err := execCommand(ctx, "git", "status", "-s")
	if err != nil {
		return errors.Wrap(err, "failed to get current branch")
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("you have unpushed changes")
	}

	out, err = execCommand(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return errors.Wrap(err, "failed to get current branch")
	}
	branch := strings.TrimSpace(out)
	if branch == master {
		return errors.New("current branch is master")
	}

	_, err = execCommand(ctx, "git", "checkout", master)
	if err != nil {
		return errors.Wrap(err, "failed to checkout master")
	}

	_, err = execCommand(ctx, "git", "pull")
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	_, err = execCommand(ctx, "git", "checkout", branch)
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	out, err = execCommand(ctx, "git", "merge", master)
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	out, err = execCommand(ctx, "git", "push")
	if err != nil {
		return errors.Wrap(err, "failed to pull")
	}

	return nil
}
