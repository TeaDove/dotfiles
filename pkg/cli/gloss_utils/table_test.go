package gloss_utils

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"testing"
)

func TestUnit_GlossUtils_NewMappingData_Ok(t *testing.T) {
	tableData := NewMappingData("name", "age", "gender")

	tableStyle := table.New().
		Wrap(true).
		Headers("name", "age", "gender").
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#df8e1d"))).
		Data(tableData)

	tableData.SetMappingRow("petya", M{"name": "petya", "age": "24", "gender": "male"})
	tableData.SetMappingRow("olya", M{"name": "olya", "age": "21", "gender": "female"})
	tableData.SetMappingRow("masha", M{"name": "masha", "age": "24", "gender": "female"})
	tableData.SetMappingRow("artem", M{"name": "artem", "age": "25", "gender": "male"})

	println(tableStyle.String())
}
