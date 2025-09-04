package cli

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

func (r *CLI) commandUUID(_ context.Context, _ *cli.Command) error {
	fmt.Print(strings.ToUpper(uuid.New().String())) //nolint:forbidigo // is ok

	return nil
}

func (r *CLI) commandText(_ context.Context, _ *cli.Command) error {
	fmt.Print(rand.Text()) //nolint:forbidigo // is ok

	return nil
}
