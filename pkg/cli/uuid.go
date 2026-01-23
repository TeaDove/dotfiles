package cli

import (
	"context"
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

func CommandUUID(_ context.Context, command *cli.Command) error {
	printStrings(command, func() string {
		return uuid.New().String()
	})

	return nil
}

func CommandText(_ context.Context, command *cli.Command) error {
	printStrings(command, rand.Text)

	return nil
}

func printStrings(command *cli.Command, fn func() string) {
	count, err := strconv.Atoi(command.Args().First())
	if err != nil || count <= 1 {
		fmt.Print(fn())
		return
	}

	var v = make([]string, 0, count)
	for range count {
		v = append(v, fn())
	}

	fmt.Print(strings.Join(v, " ")) //nolint:forbidigo // is ok
}
