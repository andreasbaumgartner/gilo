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
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func padRight(s string, w int) string {
	n := w - lipgloss.Width(s)
	if n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

// Layout helpers

const (
	listMinW = 50
	listMaxW = 100
)

func (m model) listW() int {
	if m.width <= 0 {
		return 0
	}
	w := m.width * 45 / 100
	if w < listMinW && m.width >= listMinW+20 {
		w = listMinW
	}
	if w > listMaxW {
		w = listMaxW
	}
	if w > m.width-20 {
		w = m.width - 20
	}
	if w < 0 {
		return 0
	}
	return w
}

func (m model) detailW() int    { return m.width - m.listW() }
func (m model) mainH() int      { return m.height - 2 }
func (m model) listInnerW() int { return m.listW() - 4 }

func (m model) detailInnerW() int { return m.detailW() - 4 }
func (m model) detailInnerH() int { return m.mainH() - 2 }
