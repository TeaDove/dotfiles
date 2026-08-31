package sdd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"uuid"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, cmd *cli.Command) error {
	verbose := cmd.Bool("v")

	specArg := cmd.Args().First()
	if specArg == "" {
		return errors.New("usage: dotfiles sdd <spec-file>")
	}

	specPath, err := filepath.Abs(specArg)
	if err != nil {
		return errors.Wrap(err, "resolve spec path")
	}

	content, err := os.ReadFile(specPath)
	if err != nil {
		return errors.Wrapf(err, "read spec file %s", specPath)
	}

	if strings.TrimSpace(string(content)) == "" {
		return errors.Newf("spec file %s is empty", specPath)
	}

	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return errors.Wrap(err, "claude CLI not found in PATH; install Claude Code and try again")
	}

	repoRoot := repositoryRoot(ctx)

	reporter := newConsoleReporter(verbose)
	runner := newClaudeRunner(claudeBin, repoRoot, reporter, verbose)

	printBanner(specPath, repoRoot)

	preexisting, vcsKnown := preexistingChanges(ctx, repoRoot)

	orchestrator := NewOrchestrator(runner, reporter, orchestratorConfig{
		specPath:    specPath,
		specContent: string(content),
		preexisting: preexisting,
		vcsKnown:    vcsKnown,
		session:     uuid.NewV4().String(),
	})

	err = orchestrator.Run(ctx)
	if err != nil {
		return err
	}

	color.New(color.FgHiGreen, color.Bold).Println("\n━━ SDD COMPLETE ━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

func printBanner(specPath, repoRoot string) {
	color.New(color.FgHiCyan, color.Bold).Println("\n━━ SDD ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nSpecification: %s\nRepository:    %s\n", specPath, repoRoot)
}

func repositoryRoot(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	return cwd
}

func preexistingChanges(ctx context.Context, repoRoot string) (string, bool) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoRoot

	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	return strings.TrimSpace(string(out)), true
}
