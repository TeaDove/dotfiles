package cli

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

func readFromPipeOrSTDIN() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			return "", errors.Wrap(err, "unable to read from stdin")
		}

		return text, nil
	}

	reader := bufio.NewReader(os.Stdin)
	buf := new(strings.Builder)

	_, err := io.Copy(buf, reader)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy from buf")
	}

	return buf.String(), nil
}
