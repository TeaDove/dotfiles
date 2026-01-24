package gloss_utils

import (
	"fmt"
	"slices"
	"sync"
)

type M map[string]any

type MappingData struct {
	strings [][]any

	columns []string
	rows    []string

	mu sync.RWMutex
}

func NewMappingData(columns ...string) *MappingData {
	if len(columns) == 0 {
		panic("columns must not be empty")
	}

	r := &MappingData{strings: make([][]any, 0), columns: columns, rows: make([]string, 0)}
	for range columns {
		r.strings = append(r.strings, nil)
	}

	for idx := range r.strings {
		for j := range r.strings[idx] {
			r.strings[idx][j] = ""
		}
	}

	return r
}

func (r *MappingData) Locker() sync.Locker {
	return &r.mu
}

func (r *MappingData) RLocker() sync.Locker {
	return r.mu.RLocker()
}

func (r *MappingData) addRow(row string) {
	r.rows = append(r.rows, row)
	for idx := range r.strings {
		r.strings[idx] = append(r.strings[idx], nil)
	}
}

func (r *MappingData) Set(col, row string, value any) {
	colIdx := slices.Index(r.columns, col)
	if colIdx == -1 {
		panic("column not found: " + col)
	}

	rowIdx := slices.Index(r.rows, row)
	if rowIdx == -1 {
		r.addRow(row)
		r.Set(col, row, value)

		return
	}

	r.strings[colIdx][rowIdx] = value
}

func (r *MappingData) SetMappingRow(row string, v M) {
	for col, data := range v {
		r.Set(col, row, data)
	}
}

func (r *MappingData) SetMappingColumn(col string, v M) {
	for row, data := range v {
		r.Set(col, row, data)
	}
}

func (r *MappingData) At(row, col int) string {
	if col >= len(r.strings) || row >= len(r.strings[col]) {
		return ""
	}

	v := r.strings[col][row]
	if v == nil {
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

func (r *MappingData) Clear(rows ...string) {
	notFoundRows := make([]string, 0)

	for _, row := range r.rows {
		if !slices.Contains(rows, row) {
			notFoundRows = append(notFoundRows, row)
		}
	}

	for _, row := range notFoundRows {
		r.DeleteRow(row)
	}
}

func (r *MappingData) IsEmpty() bool {
	return len(r.rows) == 0
}

func (r *MappingData) DeleteRow(row string) bool {
	idx := slices.Index(r.rows, row)
	if idx == -1 {
		return false
	}

	r.rows = slices.Delete(r.rows, idx, idx+1)
	for colIdx := range r.strings {
		r.strings[colIdx] = slices.Delete(r.strings[colIdx], idx, idx+1)
	}

	return true
}

func (r *MappingData) RowExists(row string) bool {
	return slices.Contains(r.rows, row)
}
