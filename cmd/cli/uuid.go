package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

func (r *CLI) commandUuid(ctx context.Context, command *cli.Command) error {
	fmt.Print(strings.ToUpper(uuid.New().String()))

	return nil
}
