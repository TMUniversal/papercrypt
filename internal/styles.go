package internal

import "github.com/charmbracelet/lipgloss"

var (
	// URL is used to style URLs.
	URL = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render

	// Warning is used to style warnings for the user.
	Warning = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true).Render

	BoldStyle = lipgloss.NewStyle().Bold(true)
)
