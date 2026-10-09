package tui

import "github.com/charmbracelet/lipgloss"

// Palette. AdaptiveColor picks the right shade for light and dark terminals.
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5A3FD9", Dark: "#9D8CFF"}
	colorAccent  = lipgloss.AdaptiveColor{Light: "#C2367A", Dark: "#FF7AC6"}
	colorText    = lipgloss.AdaptiveColor{Light: "#1F1F2E", Dark: "#E6E6F0"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#7A7A8C", Dark: "#7C7C94"}
	colorBorder  = lipgloss.AdaptiveColor{Light: "#D4D0E8", Dark: "#3A3654"}
	colorSuccess = lipgloss.AdaptiveColor{Light: "#1E8E5A", Dark: "#5EE6A0"}
	colorWarning = lipgloss.AdaptiveColor{Light: "#B7791F", Dark: "#FFC86B"}
	colorDanger  = lipgloss.AdaptiveColor{Light: "#C53030", Dark: "#FF6B7A"}
)

// Logo gradient, top to bottom.
var logoGradient = []lipgloss.Color{"#7C5CFF", "#9466FF", "#AC70F5", "#C47AE6", "#DC84D4", "#F08EC2"}

var (
	styleApp = lipgloss.NewStyle().Padding(1, 2)

	styleTagline   = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
	styleSignature = lipgloss.NewStyle().Foreground(colorAccent).Italic(true)
	styleGreet     = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	styleMuted     = lipgloss.NewStyle().Foreground(colorMuted)
	styleText      = lipgloss.NewStyle().Foreground(colorText)

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	stylePanelTitle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).MarginBottom(1)

	styleCardLabel = lipgloss.NewStyle().Foreground(colorMuted)
	styleCardValue = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	styleCardSub   = lipgloss.NewStyle().Foreground(colorMuted)

	styleMenuItem     = lipgloss.NewStyle().Foreground(colorText).PaddingLeft(2)
	styleMenuSelected = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	styleMenuDesc     = lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(4)
	styleKeyHint      = lipgloss.NewStyle().Foreground(colorAccent)

	stylePage       = lipgloss.NewStyle().Padding(1, 2)
	styleBrandStart = lipgloss.NewStyle().Foreground(logoGradient[0]).Bold(true)
	styleBrandEnd   = lipgloss.NewStyle().Foreground(logoGradient[len(logoGradient)-1]).Bold(true)

	styleStatusActive   = lipgloss.NewStyle().Foreground(colorSuccess)
	styleStatusPaused   = lipgloss.NewStyle().Foreground(colorWarning)
	styleStatusInactive = lipgloss.NewStyle().Foreground(colorMuted)

	styleToast = lipgloss.NewStyle().Foreground(colorWarning)
	styleError = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDanger).
			Foreground(colorDanger).
			Padding(0, 1)
)
