package main

import (
	"dotfiles/cmd/cli"

	"github.com/fatih/color"
	"github.com/teadove/teasutils/utils/logger_utils"
)

func main() {
	err := cli.NewCLI().Run(logger_utils.NewLoggedCtx())
	if err != nil {
		color.Red("Unexpected error during execution\n")
		color.White(err.Error())
	}
}
