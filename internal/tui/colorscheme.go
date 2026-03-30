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
	ReviewBadgeBg   lipgloss.Color
	ReviewBadgeFg   lipgloss.Color
	QuestionBadgeBg lipgloss.Color
	QuestionBadgeFg lipgloss.Color
	PRBadgeBg       lipgloss.Color
	PRBadgeFg       lipgloss.Color
	PRMergedBadgeBg lipgloss.Color
	PRMergedBadgeFg lipgloss.Color

	KeybindBg lipgloss.Color
	KeybindFg lipgloss.Color

	OverlayBg         lipgloss.Color
	ModalBg           lipgloss.Color
	UnfocusedSelected lipgloss.Color

	// Accent color for selected item left-border
	AccentBorder lipgloss.Color
	// Dot indicator colors for status in tab bar
	DotRunning  lipgloss.Color
	DotThinking lipgloss.Color
	DotIdle     lipgloss.Color
}

var colorSchemes = map[string]ColorScheme{
	"default":    schemeDefault,
	"catppuccin": schemeCatppuccin,
	"kanagawa":   schemeKanagawa,
}

var colorSchemeOrder = []string{"default", "catppuccin", "kanagawa"}

// Default: Modern dark theme with green accent (inspired by screenshot)
var schemeDefault = ColorScheme{
	Name: "default",

	Active:   lipgloss.Color("#50fa7b"),
	Inactive: lipgloss.Color("#44475a"),
	Dim:      lipgloss.Color("#6272a4"),
	Yellow:   lipgloss.Color("#f1fa8c"),
	Green:    lipgloss.Color("#50fa7b"),
	Red:      lipgloss.Color("#ff5555"),

	SelectedFg: lipgloss.Color("#282a36"),

	OpenBadgeBg:    lipgloss.Color("#1a3a1a"),
	OpenBadgeFg:    lipgloss.Color("#50fa7b"),
	ClosedBadgeBg:  lipgloss.Color("#3a1a1a"),
	ClosedBadgeFg:  lipgloss.Color("#ff5555"),
	WorkingBadgeBg: lipgloss.Color("#1a3a1a"),
	WorkingBadgeFg: lipgloss.Color("#50fa7b"),
	ReviewBadgeBg:   lipgloss.Color("#2a2a3a"),
	ReviewBadgeFg:   lipgloss.Color("#6272a4"),
	QuestionBadgeBg: lipgloss.Color("#3a3a1a"),
	QuestionBadgeFg: lipgloss.Color("#f1fa8c"),
	PRBadgeBg:       lipgloss.Color("#1a2a3a"),
	PRBadgeFg:       lipgloss.Color("#8be9fd"),
	PRMergedBadgeBg: lipgloss.Color("#2a1a3a"),
	PRMergedBadgeFg: lipgloss.Color("#bd93f9"),

	KeybindBg: lipgloss.Color("#282a36"),
	KeybindFg: lipgloss.Color("#f8f8f2"),

	OverlayBg:         lipgloss.Color("#1e1f29"),
	ModalBg:           lipgloss.Color("#282a36"),
	UnfocusedSelected: lipgloss.Color("#f8f8f2"),

	AccentBorder: lipgloss.Color("#50fa7b"),
	DotRunning:   lipgloss.Color("#50fa7b"),
	DotThinking:  lipgloss.Color("#f1fa8c"),
	DotIdle:      lipgloss.Color("#6272a4"),
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
	ReviewBadgeBg:   lipgloss.Color("#1a3a4a"), // Dark teal
	ReviewBadgeFg:   lipgloss.Color("#94e2d5"), // Teal
	QuestionBadgeBg: lipgloss.Color("#3a1a4a"), // Dark mauve
	QuestionBadgeFg: lipgloss.Color("#cba6f7"), // Mauve
	PRBadgeBg:       lipgloss.Color("#1a3040"), // Dark sapphire
	PRBadgeFg:       lipgloss.Color("#74c7ec"), // Sapphire
	PRMergedBadgeBg: lipgloss.Color("#2a1a3a"), // Dark lavender
	PRMergedBadgeFg: lipgloss.Color("#b4befe"), // Lavender

	KeybindBg: lipgloss.Color("#313244"), // Surface0
	KeybindFg: lipgloss.Color("#cdd6f4"), // Text

	OverlayBg:         lipgloss.Color("#11111b"), // Crust
	ModalBg:           lipgloss.Color("#1e1e2e"), // Base
	UnfocusedSelected: lipgloss.Color("#a6adc8"), // Subtext0

	AccentBorder: lipgloss.Color("#89b4fa"),
	DotRunning:   lipgloss.Color("#a6e3a1"),
	DotThinking:  lipgloss.Color("#f9e2af"),
	DotIdle:      lipgloss.Color("#6c7086"),
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
	ReviewBadgeBg:   lipgloss.Color("#1a2a3a"), // Dark blue
	ReviewBadgeFg:   lipgloss.Color("#7fb4ca"), // Spring blue
	QuestionBadgeBg: lipgloss.Color("#352a3a"), // Dark purple
	QuestionBadgeFg: lipgloss.Color("#957fb8"), // Oni violet
	PRBadgeBg:       lipgloss.Color("#1a2a3a"), // Dark wave blue
	PRBadgeFg:       lipgloss.Color("#7fb4ca"), // Spring blue
	PRMergedBadgeBg: lipgloss.Color("#2a1a30"), // Dark wisteria
	PRMergedBadgeFg: lipgloss.Color("#938aa9"), // Spring violet

	KeybindBg: lipgloss.Color("#2a2a37"), // Sumi ink 3
	KeybindFg: lipgloss.Color("#dcd7ba"), // Fuji white

	OverlayBg:         lipgloss.Color("#16161d"), // Sumi ink 0
	ModalBg:           lipgloss.Color("#1f1f28"), // Sumi ink 1
	UnfocusedSelected: lipgloss.Color("#c8c093"), // Old white

	AccentBorder: lipgloss.Color("#7e9cd8"),
	DotRunning:   lipgloss.Color("#98bb6c"),
	DotThinking:  lipgloss.Color("#e6c384"),
	DotIdle:      lipgloss.Color("#727169"),
}
