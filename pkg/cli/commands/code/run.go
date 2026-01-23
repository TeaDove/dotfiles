package code

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/tidwall/match"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func Run(_ context.Context, cmd *cli.Command) error {
	for _, arg := range cmd.Args().Slice() {
		report, err := analyzeFiles(arg)
		if err != nil {
			return errors.Wrap(err, "analyze files")
		}

		println(drawTable(&report).String())
	}

	return nil
}

func drawTable(report *Report) *table.Table {
	const (
		colTitle       = "Title"
		colCount       = "Count"
		colDescription = "Description"
	)

	other := report.TotalErrors - report.Propagates - report.PropagatesWithContext - report.Suppresses - report.Panics
	rows := []struct {
		Title       string
		Description string
		Count       uint
		Color       color.Attribute
	}{
		{
			Title:       "Total Errors",
			Description: "all err != nil",
			Count:       report.TotalErrors,
			Color:       color.FgHiWhite,
		},
		{
			Title:       "Propagates",
			Description: "if err != nil {return err}",
			Count:       report.Propagates,
			Color:       color.FgGreen,
		},
		{
			Title:       "Propagates with context",
			Description: `if err != nil {return errors.Wrap(err, ...)}`,
			Count:       report.PropagatesWithContext,
			Color:       color.FgHiGreen,
		},
		{
			Title:       "Suppresses",
			Description: "if err != nil {return nil}",
			Count:       report.Suppresses,
			Color:       color.FgYellow,
		},
		{
			Title:       "Panics",
			Description: "if err != nil {return panic(err)}",
			Count:       report.Panics,
			Color:       color.FgRed,
		},
		{
			Title:       "Other",
			Description: "anything other",
			Count:       other,
			Color:       color.FgCyan,
		},
	}

	pingCols := []string{colDescription, colTitle, colCount}
	tableData := gloss_utils.NewMappingData(pingCols...)

	for _, row := range rows {
		percent := 100 * float64(row.Count) / float64(report.TotalErrors)

		tableData.Set(colTitle, row.Title, row.Title)
		tableData.Set(
			colCount,
			row.Title,
			fmt.Sprintf("%s (%.2f%%)", color.New(row.Color).Sprintf("%d", row.Count), percent),
		)
		tableData.Set(colDescription, row.Title, row.Description)
	}

	t := table.New().
		Wrap(true).
		Headers(pingCols...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("69"))).
		Data(tableData)

	return t
}

func analyzeFile(fset *token.FileSet, filename string) (Report, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return Report{}, errors.Wrap(err, "reading source file")
	}

	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return Report{}, errors.Wrap(err, "parse file")
	}

	var report Report

	ast.Inspect(f, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
		if !ok || bin.Op != token.NEQ {
			return true
		}

		left, lOk := bin.X.(*ast.Ident)

		right, rOk := bin.Y.(*ast.Ident)
		if !lOk || !rOk || left.Name != literalErr || right.Name != literalNil {
			return true
		}

		report.TotalErrors++

		switch classify(ifStmt) {
		case errPropagate:
			color.Green("Propagate err:\n")

			report.Propagates++
		case errPropagateWithContext:
			color.HiGreen("Propagate with context err:\n")

			report.PropagatesWithContext++
		case errPanic:
			color.Red("Raise panic:\n")

			report.Panics++
		case errSuppress:
			color.Yellow("Suppress err:\n")

			report.Suppresses++
		default:
			color.Cyan("Other:\n")
		}

		pos := fset.Position(ifStmt.Pos())
		fmt.Printf("%s:%d\n", pos.Filename, pos.Line)
		fmt.Printf("%s\n\n", getIfSnippet(fset, src, ifStmt))

		return true
	})

	return report, nil
}

type Report struct {
	TotalErrors           uint
	Propagates            uint
	PropagatesWithContext uint
	Suppresses            uint
	Panics                uint
}

func analyzeFiles(path string) (Report, error) {
	fset := token.NewFileSet()

	info, err := os.Stat(path)
	if err != nil {
		return Report{}, errors.Wrap(err, "stat path")
	}

	if !info.IsDir() {
		return analyzeFile(fset, path)
	}

	var report Report

	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(p) == ".go" && !match.Match(p, "*test.go") {
			fileReport, err := analyzeFile(fset, p)
			if err != nil {
				return errors.Wrap(err, "analyze file")
			}

			report.TotalErrors += fileReport.TotalErrors
			report.Propagates += fileReport.Propagates
			report.PropagatesWithContext += fileReport.PropagatesWithContext
			report.Suppresses += fileReport.Suppresses
			report.Panics += fileReport.Panics
		}

		return nil
	})
	if err != nil {
		return Report{}, errors.Wrap(err, "walk path")
	}

	return report, nil
}
