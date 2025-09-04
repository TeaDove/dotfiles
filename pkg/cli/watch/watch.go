package watch

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
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
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
	}

	m.grepInput = textinput.New()
	m.grepInput.Cursor.SetMode(cursor.CursorHide)
	m.grepInput.Placeholder = ""
	m.grepInput.Focus()
	m.grepInput.CharLimit = 156
	m.grepInput.Width = 20

	return &Watch{model: &m}
}

var IntervalFlag = &cli.DurationFlag{ //nolint:gochecknoglobals // is ok
	Name:  "i",
	Usage: "interval between executions",
	Value: time.Second,
}

func (r *Watch) Run(ctx context.Context, cmd *cli.Command) error {
	interval := cmd.Duration(IntervalFlag.Name)

	commands := cmd.Args().Slice()
	if len(commands) == 0 {
		return errors.New("at least one command is required")
	}

	r.model.commands = make([]commandExecution, len(commands))

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

type commandExecution struct {
	cmd string
	out string
}

func (r *Watch) executeAndShow(ctx context.Context, idx int, command string, interval time.Duration) {
	var (
		iters    int
		totalDur time.Duration
		avgDur   time.Duration
	)

	for {
		t0 := time.Now()

		out, err := executeAndRead(ctx, command)
		if err != nil {
			out += color.RedString(errors.Wrap(err, "failed to run").Error())
		}

		iters++
		dur := time.Since(t0)
		totalDur += dur
		avgDur = totalDur / time.Duration(iters)

		r.model.commands[idx] = commandExecution{
			cmd: fmt.Sprintf("%s (%d), (%s)", color.MagentaString(command), iters, avgDur.String()),
			out: strings.TrimSpace(out),
		}

		time.Sleep(interval - dur)
	}
}

func executeAndRead(ctx context.Context, command string) (string, error) {
	fields := strings.Fields(command)

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...) //nolint: gosec // FP

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}
