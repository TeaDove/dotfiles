package watch

import (
	"bufio"
	"context"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Watch struct {
	model *model
}

func New() *Watch {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		spinner: s,
		help:    help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
			),
		},
	}

	return &Watch{model: &m}
}

var (
	IntervalFlag = &cli.DurationFlag{
		Name:  "i",
		Usage: "interval between executions",
		Value: time.Second,
	}
)

func (r *Watch) Run(ctx context.Context, cmd *cli.Command) error {
	interval := cmd.Duration(IntervalFlag.Name)
	commands := cmd.Args().Slice()
	if len(commands) == 0 {
		return errors.New("at least one command is required")
	}
	r.model.commands = make([]string, len(commands))

	for idx, command := range commands {
		go r.executeAndShow(ctx, idx, command, interval)
	}

	p := tea.NewProgram(r.model, tea.WithContext(ctx), tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run tea")
	}

	return nil
}

func (r *Watch) executeAndShow(ctx context.Context, idx int, command string, interval time.Duration) {
	for {
		t0 := time.Now()
		out, err := executeAndRead(ctx, command)
		if err != nil {
			out = errors.Wrap(err, "failed to run").Error()
		}

		r.model.commands[idx] = out

		time.Sleep(interval - time.Since(t0))
	}
}

func executeAndRead(ctx context.Context, command string) (string, error) {
	fields := strings.Fields(command)

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to execute command")
	}

	return string(out), nil
}

func executeAndReadChanneled(ctx context.Context, command string) <-chan string {
	fields := strings.Fields(command)

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)

	var wg sync.WaitGroup
	out := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			out <- errors.Wrap(err, "failed to open stdout pipe").Error()
			return
		}

		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run()
		if err != nil {
			out <- errors.Wrap(err, "failed to run command").Error()
		}
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
