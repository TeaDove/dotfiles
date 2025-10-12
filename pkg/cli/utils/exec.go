package utils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

func ExecCommand(ctx context.Context, name string, args ...string) (string, error) {
	color.Magenta(fmt.Sprintf("$ %s %s", name, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, name, args...)

	var outBuf bytes.Buffer

	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "execute command")
	}

	return outBuf.String(), nil
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
