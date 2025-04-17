package cli

import (
	"bufio"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"io"
	"os"
	"strings"
)

func getPipe() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", errors.New("no data in pipeline")
	}

	reader := bufio.NewReader(os.Stdin)
	buf := new(strings.Builder)

	_, err := io.Copy(buf, reader)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy from buf")
	}

	return buf.String(), nil
}

func (r *CLI) commandSha(_ context.Context, _ *cli.Command) error {
	text, err := getPipe()
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		text, err = reader.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "unable to read from stdin")
		}
	}

	hasher := sha512.New()
	hasher.Write([]byte(text))
	hashedText := hex.EncodeToString(hasher.Sum(nil))

	fmt.Print(hashedText)

	return nil
}
