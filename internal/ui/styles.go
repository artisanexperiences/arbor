package ui

import "github.com/charmbracelet/lipgloss"

var (
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#10B981")

	ColorSuccess = lipgloss.Color("#10B981")
	ColorWarning = lipgloss.Color("#F59E0B")
	ColorError   = lipgloss.Color("#EF4444")
	ColorInfo    = lipgloss.Color("#06B6D4")
	ColorMuted   = lipgloss.Color("#6B7280")

	Text    = lipgloss.Color("#F9FAFB")
	TextDim = lipgloss.Color("#9CA3AF")
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SuccessBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000")).
			Background(ColorSuccess).
			Padding(0, 1).
			Bold(true)

	WarningBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000")).
			Background(ColorWarning).
			Padding(0, 1).
			Bold(true)

	ErrorBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(ColorError).
			Padding(0, 1).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	CodeStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)

	InfoBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000")).
			Background(ColorInfo).
			Padding(0, 1).
			Bold(true)
)
