package cli

import (
	"context"
	"dotfiles/pkg/cli/utils/systemutils"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

var dotfilesDirs = [3]string{
	"./dotfiles-configs",
	"~/dotfiles/dotfiles-configs",
	"~/.dotfiles/dotfiles-configs",
}

func CommandInstall(_ context.Context, _ *cli.Command) error {
	var dofilesPath string

	for _, dir := range dotfilesDirs {
		_, err := os.Stat(dir)
		if !errors.Is(err, fs.ErrNotExist) {
			dofilesPath = dir

			break
		}
	}

	if dofilesPath == "" {
		return errors.Errorf("no dotfiles-configs dir found")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user home dir")
	}

	err = filepath.Walk(dofilesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "walk")
		}

		if info.IsDir() {
			return nil
		}

		err = os.Remove(filepath.Join(homeDir, strings.TrimPrefix(path, "dotfiles-configs/")))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errors.Wrap(err, "remove file")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "remove old files")
	}

	// TODO(teadove): some target configs must be MERGED, not overwritten.
	// e.g. ~/.claude/settings.json holds per-machine values (model, theme) that
	// differ across hosts, yet parts of it — like the Claude Code hook
	// registration that wires ~/.claude/hooks/notify.sh — should be shared via
	// dotfiles. Blindly CopyFS-ing a tracked settings.json would clobber the
	// local per-machine keys; that is why it is intentionally NOT tracked today.
	// Solution: keep an explicit list of "merge" configs and, for those, deep-
	// merge the tracked fragment into the existing file (go-json, or shell out
	// to jq: `jq -s '.[0] * .[1]'`) instead of copying, preserving local keys.
	err = os.CopyFS(homeDir, os.DirFS(dofilesPath))
	if err != nil {
		return errors.Wrap(err, "copy temp files")
	}

	color.Green("Dotfiles installed from %s to %s", dofilesPath, homeDir)

	return nil
}

func CommandUpdate(ctx context.Context, _ *cli.Command) error {
	_, err := systemutils.ExecCommand(
		ctx,
		"bash",
		"-c",
		"curl -s https://raw.githubusercontent.com/teadove/dotfiles/master/install.py | python3 -B",
	)
	if err != nil {
		return errors.Wrap(err, "install new version")
	}

	return nil
}
