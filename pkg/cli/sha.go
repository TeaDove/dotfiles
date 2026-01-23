package cli

import (
	"context"
	"crypto/sha512"
	"dotfiles/pkg/cli/utils"
	"encoding/hex"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func CommandSha(_ context.Context, cmd *cli.Command) error {
	text, err := utils.ReadFromPipeOrSTDIN()
	if err != nil {
		return errors.Wrap(err, "unable to read from stdin")
	}

	hasher := sha512.New()
	hasher.Write([]byte(text))
	hashedText := hex.EncodeToString(hasher.Sum(nil))

	if cmd.Bool(verboseFlag.Name) {
		fmt.Printf("input:\n%s\n\n", color.BlueString(text)) //nolint:forbidigo // is ok
		fmt.Printf("hash:\n%s\n", hashedText)                //nolint:forbidigo // is ok
	} else {
		fmt.Print(hashedText) //nolint:forbidigo // is ok
	}

	return nil
}
