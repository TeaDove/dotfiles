package logs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v3"
)

const (
	logDir          = "/tmp/ulogs"
	logDateFormat   = "2006-01-02"
	logExt          = ".logs"
	scannerBufStart = 64 * 1024
	scannerBufMax   = 4 * 1024 * 1024
	logDirPerm      = 0o755
	logFilePerm     = 0o644
)

func Run(ctx context.Context, _ *cli.Command) error {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return errors.Wrap(err, "open /dev/tty: this command requires an interactive terminal")
	}
	defer tty.Close()

	logPath, logFile, err := openLogFile()
	if err != nil {
		return errors.Wrap(err, "open log file")
	}
	defer logFile.Close()

	m := newModel()

	p := tea.NewProgram(m, tea.WithContext(ctx), tea.WithInput(tty))

	go streamStdin(p, logFile, logPath)

	_, err = p.Run()
	if err != nil {
		return errors.Wrap(err, "run tea")
	}

	return nil
}

func openLogFile() (string, *os.File, error) {
	err := os.MkdirAll(logDir, logDirPerm)
	if err != nil {
		return "", nil, errors.Wrap(err, "mkdir")
	}

	path := filepath.Join(logDir, time.Now().Format(logDateFormat)+logExt)

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePerm)
	if err != nil {
		return "", nil, errors.Wrap(err, "open file")
	}

	return path, file, nil
}

func streamStdin(p *tea.Program, logFile *os.File, logPath string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, scannerBufStart), scannerBufMax)

	var lineCount int

	for scanner.Scan() {
		line := scanner.Text()

		_, writeErr := fmt.Fprintln(logFile, line)
		if writeErr == nil {
			lineCount++
		}

		p.Send(logLineMsg(line))
	}

	p.Send(streamClosedMsg{path: logPath, lineCount: lineCount, err: scanner.Err()})
}
