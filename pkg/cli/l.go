package cli

import (
	"context"
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v3"
)

func CommandL(_ context.Context, _ *cli.Command) error {
	_, err := io.Copy(os.Stdout, os.Stdin)
	if err != nil {
		return errors.Wrap(err, "copy stdin to stdout")
	}

	return nil
}
