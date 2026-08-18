package main

import (
	"context"
	"dotfiles/pkg/cli"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
)

func interruptContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		cancel()
		os.Exit(int(syscall.SIGTERM))
	}()

	return ctx
}

func main() {
	err := cli.Run(interruptContext())
	if err != nil {
		color.Red("Unexpected error during execution\n")
		color.White(err.Error())
		os.Exit(1)
	}
}
