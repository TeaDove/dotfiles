package gloss_utils

import (
	"fmt"
	"slices"
)

type MappingData struct {
	strings [][]any

	columns []string
	rows    []string
}

func NewMappingData(columns []string, rows []string) *MappingData {
	r := &MappingData{strings: make([][]any, 0), columns: columns, rows: rows}
	for range columns {
		r.strings = append(r.strings, make([]any, len(rows)))
	}

	for idx := range r.strings {
		for j := range r.strings[idx] {
			r.strings[idx][j] = ""
		}
	}

	return r
}

func (r *MappingData) Set(col, row string, value any) {
	colIdx := slices.Index(r.columns, col)
	if colIdx == -1 {
		panic("column not found: " + col)
	}

	rowIdx := slices.Index(r.rows, row)
	if rowIdx == -1 {
		panic("row not found: " + row)
	}

	r.strings[colIdx][rowIdx] = value
}

func (r *MappingData) At(row, col int) string {
	if col >= len(r.strings) || row >= len(r.strings[col]) {
		return ""
	}

	return fmt.Sprintf("%v", r.strings[col][row])
}

func (r *MappingData) Rows() int {
	return len(r.rows)
}

func (r *MappingData) Columns() int {
	return len(r.columns)
}
