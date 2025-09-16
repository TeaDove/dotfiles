package cli

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

var dotfilesDirs = []string{
	"./dotfiles-configs",
	"~/dotfiles/dotfiles-configs",
	"~/.dotfiles/dotfiles-configs",
} //nolint:gochecknoglobals // is ok

func (r *CLI) commandInstall(_ context.Context, _ *cli.Command) error {
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

	err = os.CopyFS(homeDir, os.DirFS(dofilesPath))
	if err != nil {
		return errors.Wrap(err, "copy temp files")
	}

	color.Green("Dotfiles installed from %s to %s", dofilesPath, homeDir)

	return nil
}

func (r *CLI) commandUpdate(ctx context.Context, _ *cli.Command) error {
	_, err := utils.ExecCommand(
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
