package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func downloadDotfiles() (*zip.Reader, error) {
	resp, err := http.Get( //nolint: noctx // don't care
		"https://github.com/TeaDove/dotfiles/archive/refs/heads/master.zip",
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download dotfiles")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}
	defer resp.Body.Close()

	zipArchive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open zip archive")
	}

	color.White("Dotfiles downloaded")

	return zipArchive, nil
}

func (r *CLI) commandInstall(_ context.Context, _ *cli.Command) error {
	zipArchive, err := downloadDotfiles()
	if err != nil {
		return errors.Wrap(err, "failed to download dotfiles")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "failed to get user home dir")
	}

	var count int

	for _, file := range zipArchive.File {
		if file.FileInfo().IsDir() || !strings.HasPrefix(file.Name, "dotfiles-master/dotfiles-configs/") {
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}

		fileContent, err := io.ReadAll(fileReader)
		if err != nil {
			return errors.Wrap(err, "failed to read file")
		}

		name := strings.TrimPrefix(file.Name, "dotfiles-master/dotfiles-configs/")
		targetName := homeDir + "/" + name

		err = os.WriteFile(targetName, fileContent, file.Mode())
		if err != nil {
			return errors.Wrap(err, "failed to write file")
		}

		count++
	}

	color.Green("Files loaded: %d", count)

	return nil
}
