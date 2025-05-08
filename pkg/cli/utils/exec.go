package utils

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ExecCommand(ctx context.Context, name string, args ...string) (string, error) {
	color.Magenta(fmt.Sprintf("$ %s %s", name, strings.Join(args, " ")))
	cmd := exec.CommandContext(ctx, name, args...)

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current branch")
	}

	color.White(string(out))

	return string(out), nil
}

func ReadFromPipeOrSTDIN() (string, error) {
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
