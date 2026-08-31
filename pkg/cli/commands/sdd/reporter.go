package sdd

import (
	"fmt"

	"github.com/fatih/color"
)

type Reporter interface {
	Stage(name string)
	Info(message string)
	Tool(name, detail string)
	Success(message string)
	Warning(message string)
	Error(message string)
	Verbose(message string)
}

type consoleReporter struct {
	verbose bool
}

func newConsoleReporter(verbose bool) *consoleReporter {
	return &consoleReporter{verbose: verbose}
}

func (r *consoleReporter) Stage(name string) {
	color.New(color.FgHiCyan, color.Bold).Printf("\n[%s]\n\n", name)
}

func (r *consoleReporter) Info(message string) {
	fmt.Printf("  %s\n", message)
}

func (r *consoleReporter) Tool(name, detail string) {
	switch name {
	case toolRead:
		fmt.Printf("  %s %s\n", color.HiBlackString("read"), detail)
	case toolEdit, toolWrite, toolMultiEdit, toolNotebookEdit:
		fmt.Printf("  %s %s\n", color.HiBlackString("edit"), detail)
	case toolBash:
		color.Magenta("  $ %s", detail)
	default:
		if r.verbose {
			fmt.Printf("  %s %s(%s)\n", color.HiBlackString("tool"), name, detail)
		}
	}
}

func (r *consoleReporter) Success(message string) {
	color.Green("  ✓ %s", message)
}

func (r *consoleReporter) Warning(message string) {
	color.Yellow("  ! %s", message)
}

func (r *consoleReporter) Error(message string) {
	color.Red("  ✗ %s", message)
}

func (r *consoleReporter) Verbose(message string) {
	if !r.verbose {
		return
	}

	fmt.Printf("  %s %s\n", color.HiBlackString("»"), message)
}
