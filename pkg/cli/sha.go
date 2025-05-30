package cli

import (
	"context"
	"crypto/sha512"
	"dotfiles/pkg/cli/utils"
	"encoding/hex"
	"fmt"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func (r *CLI) commandSha(_ context.Context, cmd *cli.Command) error {
	text, err := utils.ReadFromPipeOrSTDIN()
	if err != nil {
		return errors.Wrap(err, "unable to read from stdin")
	}

	hasher := sha512.New()
	hasher.Write([]byte(text))
	hashedText := hex.EncodeToString(hasher.Sum(nil))

	if cmd.Bool(verboseFlag.Name) {
		fmt.Printf("input:\n%s\n\n", color.BlueString(text))
		fmt.Printf("hash:\n%s\n", hashedText)
	} else {
		fmt.Print(hashedText)
	}

	return nil
}
