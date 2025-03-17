package cli

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
	"strings"
)

func (r *CLI) commandUuid(ctx context.Context, command *cli.Command) error {
	fmt.Print(strings.ToUpper(uuid.New().String()))

	return nil
}
