package main

import (
	"context"
	"dotfiles/pkg/cli"

	"github.com/fatih/color"
)

func main() {
	err := cli.NewCLI().Run(context.Background())
	if err != nil {
		color.Red("Unexpected error during execution\n")
		color.White(err.Error())
	}
}
