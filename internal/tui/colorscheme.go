package tui

import "github.com/charmbracelet/lipgloss"

// ColorScheme defines all colors used by the TUI.
type ColorScheme struct {
	Name string

	Active   lipgloss.Color
	Inactive lipgloss.Color
	Dim      lipgloss.Color
	Yellow   lipgloss.Color
	Green    lipgloss.Color
	Red      lipgloss.Color

	SelectedFg lipgloss.Color

	OpenBadgeBg    lipgloss.Color
	OpenBadgeFg    lipgloss.Color
	ClosedBadgeBg  lipgloss.Color
	ClosedBadgeFg  lipgloss.Color
	WorkingBadgeBg lipgloss.Color
	WorkingBadgeFg lipgloss.Color
	ReviewBadgeBg  lipgloss.Color
	ReviewBadgeFg  lipgloss.Color

	KeybindBg lipgloss.Color
	KeybindFg lipgloss.Color

	OverlayBg         lipgloss.Color
	UnfocusedSelected lipgloss.Color
}

var colorSchemes = map[string]ColorScheme{
	"default":    schemeDefault,
	"catppuccin": schemeCatppuccin,
	"kanagawa":   schemeKanagawa,
}

var colorSchemeOrder = []string{"default", "catppuccin", "kanagawa"}

// Default: Lazygit-inspired clean white theme
var schemeDefault = ColorScheme{
	Name: "default",

	Active:   lipgloss.Color("62"),
	Inactive: lipgloss.Color("240"),
	Dim:      lipgloss.Color("240"),
	Yellow:   lipgloss.Color("3"),
	Green:    lipgloss.Color("2"),
	Red:      lipgloss.Color("1"),

	SelectedFg: lipgloss.Color("15"),

	OpenBadgeBg:    lipgloss.Color("22"),
	OpenBadgeFg:    lipgloss.Color("10"),
	ClosedBadgeBg:  lipgloss.Color("52"),
	ClosedBadgeFg:  lipgloss.Color("9"),
	WorkingBadgeBg: lipgloss.Color("58"),
	WorkingBadgeFg: lipgloss.Color("11"),
	ReviewBadgeBg:  lipgloss.Color("23"),
	ReviewBadgeFg:  lipgloss.Color("14"),

	KeybindBg: lipgloss.Color("236"),
	KeybindFg: lipgloss.Color("252"),

	OverlayBg:         lipgloss.Color("0"),
	UnfocusedSelected: lipgloss.Color("252"),
}

// Catppuccin Mocha
var schemeCatppuccin = ColorScheme{
	Name: "catppuccin",

	Active:   lipgloss.Color("#89b4fa"), // Blue
	Inactive: lipgloss.Color("#585b70"), // Surface2
	Dim:      lipgloss.Color("#6c7086"), // Overlay0
	Yellow:   lipgloss.Color("#f9e2af"), // Yellow
	Green:    lipgloss.Color("#a6e3a1"), // Green
	Red:      lipgloss.Color("#f38ba8"), // Red

	SelectedFg: lipgloss.Color("#1e1e2e"), // Base (dark bg)

	OpenBadgeBg:    lipgloss.Color("#1e4620"), // Dark green
	OpenBadgeFg:    lipgloss.Color("#a6e3a1"), // Green
	ClosedBadgeBg:  lipgloss.Color("#4e1a2a"), // Dark red
	ClosedBadgeFg:  lipgloss.Color("#f38ba8"), // Red
	WorkingBadgeBg: lipgloss.Color("#4a4020"), // Dark yellow
	WorkingBadgeFg: lipgloss.Color("#f9e2af"), // Yellow
	ReviewBadgeBg:  lipgloss.Color("#1a3a4a"), // Dark teal
	ReviewBadgeFg:  lipgloss.Color("#94e2d5"), // Teal

	KeybindBg: lipgloss.Color("#313244"), // Surface0
	KeybindFg: lipgloss.Color("#cdd6f4"), // Text

	OverlayBg:         lipgloss.Color("#11111b"), // Crust
	UnfocusedSelected: lipgloss.Color("#a6adc8"), // Subtext0
}

// Kanagawa (inspired by the famous Neovim theme)
var schemeKanagawa = ColorScheme{
	Name: "kanagawa",

	Active:   lipgloss.Color("#7e9cd8"), // Crystal blue
	Inactive: lipgloss.Color("#54546d"), // Sumi ink 4
	Dim:      lipgloss.Color("#727169"), // Fuji gray
	Yellow:   lipgloss.Color("#e6c384"), // Carp yellow
	Green:    lipgloss.Color("#98bb6c"), // Spring green
	Red:      lipgloss.Color("#c34043"), // Autumn red

	SelectedFg: lipgloss.Color("#1f1f28"), // Sumi ink 1

	OpenBadgeBg:    lipgloss.Color("#2a3a2a"), // Dark green
	OpenBadgeFg:    lipgloss.Color("#98bb6c"), // Spring green
	ClosedBadgeBg:  lipgloss.Color("#3a2020"), // Dark red
	ClosedBadgeFg:  lipgloss.Color("#c34043"), // Autumn red
	WorkingBadgeBg: lipgloss.Color("#3a3520"), // Dark yellow
	WorkingBadgeFg: lipgloss.Color("#e6c384"), // Carp yellow
	ReviewBadgeBg:  lipgloss.Color("#1a2a3a"), // Dark blue
	ReviewBadgeFg:  lipgloss.Color("#7fb4ca"), // Spring blue

	KeybindBg: lipgloss.Color("#2a2a37"), // Sumi ink 3
	KeybindFg: lipgloss.Color("#dcd7ba"), // Fuji white

	OverlayBg:         lipgloss.Color("#16161d"), // Sumi ink 0
	UnfocusedSelected: lipgloss.Color("#c8c093"), // Old white
}
