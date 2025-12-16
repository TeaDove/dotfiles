package main

import (
	"context"
	"dotfiles/pkg/cli"
	"os"

	"github.com/fatih/color"
)

func main() {
	// file, err := os.OpenFile(".profile.pprof", os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	//	panic(err)
	//}
	// defer file.Close()
	//
	// err = pprof.StartCPUProfile(file)
	// if err != nil {
	//	panic(err)
	//}
	// defer pprof.StopCPUProfile()
	err := cli.Run(context.Background())
	if err != nil {
		color.Red("Unexpected error during execution\n")
		color.White(err.Error())
		os.Exit(1)
	}
}
