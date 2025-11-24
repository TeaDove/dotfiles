package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/tidwall/match"
	"github.com/urfave/cli/v3"
)

func CommandStarshipSwap(_ context.Context, _ *cli.Command) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "user home dir")
	}

	const (
		starshipPattern = "starship*.toml"
		starshipConfig  = "starship.toml"
	)

	var (
		starshipBasePath   = fmt.Sprintf("%s/.config", homedir)
		starshipConfigPath = fmt.Sprintf("%s/starship.toml", starshipBasePath)
		starshipSwapPath   = fmt.Sprintf("%s/starship-swap.txt", starshipBasePath)
	)

	swapIndex, err := getSwapIdx(starshipSwapPath)
	if err != nil {
		return errors.Wrap(err, "get swap idx")
	}

	paths, err := os.ReadDir(starshipBasePath)
	if err != nil {
		return errors.Wrap(err, "read starship dir")
	}

	var configs []string

	for _, path := range paths {
		if !match.Match(path.Name(), starshipPattern) || path.Name() == starshipConfig {
			continue
		}

		configs = append(configs, path.Name())
	}

	swapIndex = (swapIndex + 1) % len(configs)

	err = setSwap(starshipSwapPath, swapIndex)
	if err != nil {
		return errors.Wrap(err, "set swap")
	}

	newConfigFile := fmt.Sprintf("%s/%s", starshipBasePath, configs[swapIndex])

	err = copyFile(newConfigFile, starshipConfigPath)
	if err != nil {
		return errors.Wrap(err, "copy file")
	}

	fmt.Printf("Configs: \n")

	for idx, config := range configs {
		msg := fmt.Sprintf("  (%d) %s\n", idx, config)

		if idx == swapIndex {
			color.Cyan(msg)
		} else {
			fmt.Print(msg)
		}
	}

	return nil
}

func getSwapIdx(starshipSwapPath string) (int, error) {
	swapContent, err := os.ReadFile(starshipSwapPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return 0, errors.Wrap(err, "starship swap readfile")
		}

		swapContent = []byte("0")

		err = setSwap(starshipSwapPath, 0)
		if err != nil {
			return 0, errors.Wrap(err, "set swap")
		}
	}

	swapIndex, err := strconv.Atoi(string(bytes.TrimSpace(swapContent)))
	if err != nil {
		swapIndex = 0

		err = setSwap(starshipSwapPath, 0)
		if err != nil {
			return 0, errors.Wrap(err, "set swap")
		}
	}

	return swapIndex, nil
}

func setSwap(starshipSwapPath string, idx int) error {
	err := os.WriteFile(starshipSwapPath, []byte(strconv.Itoa(idx)), 0600)
	if err != nil {
		return errors.Wrap(err, "starship swap init writefile")
	}

	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open source file")
	}

	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "create dest file")
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrap(err, "copy file")
	}

	return nil
}
