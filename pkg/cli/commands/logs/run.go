package logs

import (
	"bufio"
	"context"
	"dotfiles/pkg/cli/utils/closeutils"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

const logDir = "/tmp/ulog"

//nolint:gochecknoglobals // is ok
var NoSaveFlag = &cli.BoolFlag{Name: "no-save", Usage: "do not save log to /tmp/ulog"}

func Run(ctx context.Context, cmd *cli.Command) error {
	noSave := cmd.Bool(NoSaveFlag.Name)

	formatter, err := NewLogFormatter(cmd.Bool("v"))
	if err != nil {
		return errors.Wrap(err, "new log formatter")
	}

	var (
		logWriter io.Writer
		logPath   string
	)

	if !noSave {
		mkdirErr := os.MkdirAll(logDir, 0o755)
		if mkdirErr != nil {
			return errors.Wrap(mkdirErr, "make log dir")
		}

		logPath = filepath.Join(logDir, time.Now().Format("2006-01-02-15-04")+".txt")

		createdFile, createErr := os.Create(logPath)
		if createErr != nil {
			return errors.Wrap(createErr, "create log file")
		}
		defer closeutils.MustClose(createdFile)

		logWriter = createdFile
	}

	done := make(chan error, 1)

	go func() {
		done <- colorize(formatter, os.Stdin, os.Stdout, logWriter)
	}()

	select {
	case err = <-done:
	case <-ctx.Done():
		err = ctx.Err()
	}

	if !noSave {
		color.HiCyan("\nstdout saved to %s\n", logPath)
	}

	if err != nil {
		return errors.Wrap(err, "colorize")
	}

	return nil
}

func colorize(formatter *LogFormatter, reader io.Reader, stdout io.Writer, logFile io.Writer) error {
	bufReader := bufio.NewReader(reader)

	for {
		line, err := bufReader.ReadString('\n')
		if len(line) > 0 {
			if logFile != nil {
				_, writeErr := io.WriteString(logFile, line) //nolint: gosec // no taint
				if writeErr != nil {
					return errors.Wrap(writeErr, "write to log file")
				}
			}

			_, writeErr := io.WriteString(stdout, formatter.format(line)) //nolint: gosec // no taint
			if writeErr != nil {
				return errors.Wrap(writeErr, "write to stdout")
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return errors.Wrap(err, "read line")
		}
	}
}
