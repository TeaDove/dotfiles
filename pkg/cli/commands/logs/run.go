package logs

import (
	"bufio"
	"context"
	"dotfiles/pkg/cli/utils/closeutils"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

const logDir = "/tmp/ulog"

var NoSaveFlag = &cli.BoolFlag{Name: "no-save", Usage: "do not save log to /tmp/ulog"} //nolint:gochecknoglobals // is ok

var logLineRegexp = regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2} (\d{2}:\d{2}:\d{2})\.\d+ (\[[^\]]*\]) ([A-Z]): (.*)$`) //nolint:gochecknoglobals,lll // is ok

var levelColors = map[string]color.Attribute{ //nolint:gochecknoglobals // is ok
	"D": color.FgHiBlack,
	"I": color.FgGreen,
	"W": color.FgYellow,
	"E": color.FgRed,
	"P": color.FgHiRed,
	"F": color.FgHiRed,
}

func Run(_ context.Context, cmd *cli.Command) error {
	noSave := cmd.Bool(NoSaveFlag.Name)

	var (
		logWriter io.Writer
		logPath   string
	)

	if !noSave {
		err := os.MkdirAll(logDir, 0o755)
		if err != nil {
			return errors.Wrap(err, "make log dir")
		}

		logPath = filepath.Join(logDir, time.Now().Format("2006-01-02-15-04")+".txt")

		logFile, err := os.Create(logPath)
		if err != nil {
			return errors.Wrap(err, "create log file")
		}
		defer closeutils.MustClose(logFile)

		logWriter = logFile
	}

	err := colorize(os.Stdin, os.Stdout, logWriter)
	if err != nil {
		return errors.Wrap(err, "colorize")
	}

	if !noSave {
		color.HiCyan("\nstdout saved to %s\n", logPath)
	}

	return nil
}

func colorize(reader io.Reader, stdout io.Writer, logFile io.Writer) error {
	bufReader := bufio.NewReader(reader)

	for {
		line, err := bufReader.ReadString('\n')
		if len(line) > 0 {
			if logFile != nil {
				_, writeErr := io.WriteString(logFile, line)
				if writeErr != nil {
					return errors.Wrap(writeErr, "write to log file")
				}
			}

			_, writeErr := io.WriteString(stdout, prettifyLine(line))
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

func prettifyLine(line string) string {
	trimmed := strings.TrimSuffix(line, "\n")

	matches := logLineRegexp.FindStringSubmatch(trimmed)
	if matches == nil {
		return line
	}

	timePart, caller, level, message := matches[1], matches[2], matches[3], matches[4]

	levelColor, ok := levelColors[level]
	if !ok {
		return line
	}

	suffix := ""
	if strings.HasSuffix(line, "\n") {
		suffix = "\n"
	}

	coloredCaller := color.New(color.FgHiCyan).Sprint(caller)
	coloredLevel := color.New(levelColor).Sprintf("%s:", level)

	// TODO(teadove): через fmt.Sprintf
	return timePart + " " + coloredCaller + " " + coloredLevel + " " + message + suffix
}
