package display

import "github.com/charmbracelet/lipgloss"

// Colors for the TUI
var (
	primaryColor   = lipgloss.Color("#00FFFF") // Cyan
	secondaryColor = lipgloss.Color("#FFFF00") // Yellow
	accentColor    = lipgloss.Color("#00FF00") // Green
	dimColor       = lipgloss.Color("#666666") // Gray
	rootColor      = lipgloss.Color("#FF6666") // Red for root notes
)

// Base styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	chordStyle = lipgloss.NewStyle().
			Width(20).
			Align(lipgloss.Left)

	currentChordStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				Width(20).
				Align(lipgloss.Left)

	lyricsStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Width(20)

	beatStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	currentBeatStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	columnStyle = lipgloss.NewStyle().
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#444444"))

	progressStyle = lipgloss.NewStyle().
			Foreground(accentColor)
)
