package tui

import "github.com/charmbracelet/lipgloss"

var activeScheme ColorScheme

var (
	colorActive   lipgloss.Color
	colorInactive lipgloss.Color
	colorDim      lipgloss.Color
	colorYellow   lipgloss.Color
	colorGreen    lipgloss.Color
	colorRed      lipgloss.Color

	activeBorder   lipgloss.Style
	inactiveBorder lipgloss.Style
	titleStyle     lipgloss.Style
	selectedStyle  lipgloss.Style
	dimStyle       lipgloss.Style
	yellowStyle    lipgloss.Style
	greenStyle     lipgloss.Style
	redStyle       lipgloss.Style
	openBadge      lipgloss.Style
	closedBadge    lipgloss.Style
	workingBadge   lipgloss.Style
	questionBadge  lipgloss.Style
	reviewBadge    lipgloss.Style
	keybindStyle   lipgloss.Style
	modalStyle     lipgloss.Style

	// New styles for design refresh
	accentBorder  lipgloss.Style
	dotRunning    lipgloss.Style
	dotThinking   lipgloss.Style
	dotIdle       lipgloss.Style
)

func init() {
	applyColorScheme(schemeDefault)
}

func applyColorScheme(cs ColorScheme) {
	activeScheme = cs

	colorActive = cs.Active
	colorInactive = cs.Inactive
	colorDim = cs.Dim
	colorYellow = cs.Yellow
	colorGreen = cs.Green
	colorRed = cs.Red

	activeBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorActive)

	inactiveBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorInactive)

	titleStyle = lipgloss.NewStyle().
		Foreground(colorActive).
		Bold(true)

	selectedStyle = lipgloss.NewStyle().
		Foreground(colorActive).
		Bold(true)

	dimStyle = lipgloss.NewStyle().Foreground(colorDim)
	yellowStyle = lipgloss.NewStyle().Foreground(colorYellow)
	greenStyle = lipgloss.NewStyle().Foreground(colorGreen)
	redStyle = lipgloss.NewStyle().Foreground(colorRed)

	openBadge = lipgloss.NewStyle().
		Background(cs.OpenBadgeBg).
		Foreground(cs.OpenBadgeFg).
		Bold(true).
		Padding(0, 1)

	closedBadge = lipgloss.NewStyle().
		Background(cs.ClosedBadgeBg).
		Foreground(cs.ClosedBadgeFg).
		Bold(true).
		Padding(0, 1)

	workingBadge = lipgloss.NewStyle().
		Background(cs.WorkingBadgeBg).
		Foreground(cs.WorkingBadgeFg).
		Padding(0, 1)

	questionBadge = lipgloss.NewStyle().
		Background(cs.QuestionBadgeBg).
		Foreground(cs.QuestionBadgeFg).
		Padding(0, 1)

	reviewBadge = lipgloss.NewStyle().
		Background(cs.ReviewBadgeBg).
		Foreground(cs.ReviewBadgeFg).
		Padding(0, 1)

	keybindStyle = lipgloss.NewStyle().
		Background(cs.KeybindBg).
		Foreground(cs.KeybindFg).
		Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorActive).
		Padding(1, 2).
		Width(54)

	// Accent border for selected items (green left bar)
	accentBorder = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(cs.AccentBorder).
		PaddingLeft(1)

	// Colored dot styles for tab bar indicators
	dotRunning = lipgloss.NewStyle().Foreground(cs.DotRunning)
	dotThinking = lipgloss.NewStyle().Foreground(cs.DotThinking)
	dotIdle = lipgloss.NewStyle().Foreground(cs.DotIdle)
}
