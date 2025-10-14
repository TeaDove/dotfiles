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

func getLen(command *cli.Command) uint {
	arg, err := strconv.Atoi(command.Args().First())
	if err != nil {
		return 1
	}

	return uint(arg)
}

func CommandUUID(_ context.Context, command *cli.Command) error {
	arg := getLen(command)
	for range arg {
		fmt.Print(strings.ToUpper(uuid.New().String())) //nolint:forbidigo // is ok
	}

	return nil
}

func CommandText(_ context.Context, command *cli.Command) error {
	arg := getLen(command)
	for range arg {
		fmt.Print(rand.Text()) //nolint:forbidigo // is ok
	}

	return nil
}
