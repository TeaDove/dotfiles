package logs

import (
	"bufio"
	"context"
	"dotfiles/pkg/cli/utils/closeutils"
	"fmt"
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

type Log struct {
	time    string
	caller  string
	level   string
	message string
}

type LogParser struct {
	lineRegexp  *regexp.Regexp
	tagRegexp   *regexp.Regexp
	levelColors map[string]color.Attribute
	callerColor color.Attribute
	tagColor    color.Attribute
}

func NewLogParser() *LogParser {
	return &LogParser{
		lineRegexp: regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2} (\d{2}:\d{2}:\d{2})\.\d+ (\[[^\]]*\]) ([A-Z]): (.*)$`),
		tagRegexp:  regexp.MustCompile(`^(\[[^\]]*=[^\]]*\])(\s*)`),
		levelColors: map[string]color.Attribute{
			"D": color.FgHiBlack,
			"I": color.FgGreen,
			"W": color.FgYellow,
			"E": color.FgRed,
			"P": color.FgHiRed,
			"F": color.FgHiRed,
		},
		callerColor: color.FgHiCyan,
		tagColor:    color.FgHiBlack,
	}
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

	err := NewLogParser().colorize(os.Stdin, os.Stdout, logWriter)
	if err != nil {
		return errors.Wrap(err, "colorize")
	}

	if !noSave {
		color.HiCyan("\nstdout saved to %s\n", logPath)
	}

	return nil
}

func (parser *LogParser) colorize(reader io.Reader, stdout io.Writer, logFile io.Writer) error {
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

			_, writeErr := io.WriteString(stdout, parser.prettifyLine(line))
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

func (parser *LogParser) prettifyLine(line string) string {
	trimmed := strings.TrimSuffix(line, "\n")

	parsedLog, ok := parser.parse(trimmed)
	if !ok {
		return line
	}

	levelColor, ok := parser.levelColors[parsedLog.level]
	if !ok {
		return line
	}

	suffix := ""
	if strings.HasSuffix(line, "\n") {
		suffix = "\n"
	}

	coloredCaller := color.New(parser.callerColor).Sprint(parsedLog.caller)
	coloredLevel := color.New(levelColor).Sprintf("%s:", parsedLog.level)
	coloredMessage := parser.colorizeTags(parsedLog.message)

	return fmt.Sprintf("%s %s %s %s%s", parsedLog.time, coloredCaller, coloredLevel, coloredMessage, suffix)
}

func (parser *LogParser) parse(line string) (Log, bool) {
	matches := parser.lineRegexp.FindStringSubmatch(line)
	if matches == nil {
		return Log{}, false
	}

	return Log{time: matches[1], caller: matches[2], level: matches[3], message: matches[4]}, true
}

func (parser *LogParser) colorizeTags(message string) string {
	var builder strings.Builder

	rest := message
	for {
		matches := parser.tagRegexp.FindStringSubmatch(rest)
		if matches == nil {
			break
		}

		builder.WriteString(color.New(parser.tagColor).Sprint(matches[1]))
		builder.WriteString(matches[2])
		rest = rest[len(matches[0]):]
	}

	builder.WriteString(rest)

	return builder.String()
}
