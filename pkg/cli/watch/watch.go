package watch

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Watch struct{}

func New() *Watch {
	return &Watch{}
}

var (
	IntervalFlag = &cli.DurationFlag{
		Name:  "i",
		Usage: "interval between executions",
		Value: time.Second,
	}
)

func (r *Watch) Run(ctx context.Context, cmd *cli.Command) error {
	//interval := cmd.Duration(IntervalFlag.Name)
	commands := cmd.Args().Slice()
	if len(commands) == 0 {
		return errors.New("at least one command is required")
	}

	var commandChannels []<-chan string
	for _, command := range commands {
		commandChannels = append(commandChannels, executeAndRead(ctx, command))
	}

	for _, channel := range commandChannels {
		for line := range channel {
			fmt.Println(line)
		}
	}

	return nil
}

func executeAndRead(ctx context.Context, command string) <-chan string {
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
