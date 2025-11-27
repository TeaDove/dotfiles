package code

import (
	"go/ast"
	"go/token"
	"math"
	"slices"
	"strings"
	"unicode"
)

func getIfSnippet(fset *token.FileSet, src []byte, ifStmt *ast.IfStmt) string {
	startPos := fset.Position(ifStmt.Pos())
	endPos := fset.Position(ifStmt.End())

	// используем byte-Offsets для вытаскивания исходника
	start := startPos.Offset
	end := endPos.Offset

	if start < 0 {
		start = 0
	}

	if end > len(src) {
		end = len(src)
	}

	if start >= end {
		return ""
	}

	snippet := src[start:end]

	return squeezeLines(string(snippet))
}

func countLeftSpaces(line string) int {
	var spaces int

	for idx, char := range line {
		if !unicode.IsSpace(char) {
			break
		}

		spaces = idx + 1
	}

	return spaces
}

func getMinSpaces(lines []string) int {
	var minSpaces = math.MaxInt

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		spaces := countLeftSpaces(line)

		if spaces < minSpaces {
			minSpaces = spaces
		}
	}

	return minSpaces
}

func squeezeLines(src string) string {
	var builder strings.Builder

	lines := slices.Collect(strings.Lines(src))
	if len(lines) <= 1 {
		return src
	}

	builder.WriteString(lines[0])

	maxSpaces := getMinSpaces(lines[1:])

	for _, line := range lines[1:] {
		var startsAt int

		for idx, char := range line {
			if !unicode.IsSpace(char) || idx >= maxSpaces {
				break
			}

			startsAt = idx
		}

		builder.WriteString(line[startsAt+1:])
	}

	return builder.String()
}
