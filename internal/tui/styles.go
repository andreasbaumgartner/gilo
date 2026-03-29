package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorActive   = lipgloss.Color("62")
	colorInactive = lipgloss.Color("240")
	colorDim      = lipgloss.Color("240")
	colorYellow   = lipgloss.Color("3")
	colorGreen    = lipgloss.Color("2")
	colorRed      = lipgloss.Color("1")

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
			Background(colorActive).
			Foreground(lipgloss.Color("15")).
			Bold(true)

	dimStyle    = lipgloss.NewStyle().Foreground(colorDim)
	yellowStyle = lipgloss.NewStyle().Foreground(colorYellow)
	greenStyle  = lipgloss.NewStyle().Foreground(colorGreen)
	redStyle    = lipgloss.NewStyle().Foreground(colorRed)

	openBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("22")).
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Padding(0, 1)

	closedBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Padding(0, 1)

	workingBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("58")).
			Foreground(lipgloss.Color("11")).
			Bold(true).
			Padding(0, 1)

	reviewBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("23")).
			Foreground(lipgloss.Color("14")).
			Bold(true).
			Padding(0, 1)

	keybindStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorActive).
			Padding(1, 2).
			Width(54)
)
