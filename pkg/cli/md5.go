package cli

import (
	"context"
	"crypto/md5" //nolint: gosec // as expected
	"dotfiles/pkg/cli/utils"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

func CommandMD5UUID(_ context.Context, cmd *cli.Command) error {
	text, err := utils.ReadFromPipeOrSTDIN()
	if err != nil {
		return errors.Wrap(err, "unable to read from stdin")
	}

	hash := md5.Sum([]byte(text)) //nolint: gosec // as expected
	hashedText := uuid.UUID(hash[:]).String()

	if cmd.Bool(verboseFlag.Name) {
		fmt.Printf("input:\n%s\n\n", color.BlueString(text)) //nolint:forbidigo // is ok
		fmt.Printf("hash:\n%s\n", hashedText)                //nolint:forbidigo // is ok
	} else {
		fmt.Print(hashedText) //nolint:forbidigo // is ok
	}

	return nil
}
