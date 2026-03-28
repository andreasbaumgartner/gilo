package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func panelBorder(active bool) lipgloss.Style {
	if active {
		return activeBorder
	}
	return inactiveBorder
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func padRight(s string, w int) string {
	n := w - lipgloss.Width(s)
	if n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

// Layout helpers

func (m model) listW() int      { return m.width * 38 / 100 }
func (m model) detailW() int    { return m.width - m.listW() }
func (m model) mainH() int      { return m.height - 2 }
func (m model) listInnerW() int { return m.listW() - 4 }

func (m model) detailInnerW() int { return m.detailW() - 4 }
func (m model) detailInnerH() int { return m.mainH() - 2 }
