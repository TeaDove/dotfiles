package cli

import (
	"context"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	dotfilesDirs = []string{"./dotfiles-configs", "~/dotfiles/dotfiles-configs", "~/.dotfiles/dotfiles-configs"}
)

const (
	dest = ".temp"
)

func (r *CLI) commandInstall(ctx context.Context, command *cli.Command) error {
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

	err := filepath.Walk(dofilesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to walk")
		}
		if info.IsDir() {
			return nil
		}

		err = os.Remove(filepath.Join(dest, strings.TrimPrefix(path, "dotfiles-configs/")))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errors.Wrap(err, "failed to remove file")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to remove old files")
	}

	err = os.CopyFS(dest, os.DirFS(dofilesPath))
	if err != nil {

		return errors.Wrap(err, "failed to copy temp files")
	}

	return nil
}
