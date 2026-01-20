package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func RenderTable(headers []string, rows [][]string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(Primary)).
		Headers(headers...)

	for _, row := range rows {
		t.Row(row...)
	}

	return t.String()
}

func RenderStatusTable(rows [][]string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(Primary)).
		Headers("TOOL", "STATUS", "VERSION").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(Primary)
			}
			if col == 1 {
				return lipgloss.NewStyle().Foreground(ColorSuccess)
			}
			return lipgloss.Style{}
		})

	for _, row := range rows {
		t.Row(row...)
	}

	return fmt.Sprintf("\n%s\n", t.String())
}
