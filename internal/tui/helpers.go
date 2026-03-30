package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	if m.listWidthOverride > 0 {
		w := m.listWidthOverride
		if w < listMinW {
			w = listMinW
		}
		if w > m.width-20 {
			w = m.width - 20
		}
		return w
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

// modalW returns a responsive modal width based on terminal size.
func (m model) modalW() int {
	w := m.width * 50 / 100
	if w < 50 {
		w = 50
	}
	if w > 70 {
		w = 70
	}
	if w > m.width-4 {
		w = m.width - 4
	}
	return w
}

// textareaW returns the inner width available for textareas inside a modal.
func (m model) textareaW() int {
	// modal border (1 each side) + padding (2 each side) = 6
	w := m.modalW() - 6
	if w < 20 {
		w = 20
	}
	return w
}

// overlayCenter renders the modal centered over a dimmed version of the
// background, creating a "blurred" popup effect.
func overlayCenter(bg string, modal string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	dimFg := lipgloss.NewStyle().Foreground(activeScheme.OverlayBg)

	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}
	bgLines = bgLines[:height]

	// Strip ANSI from background to get plain text
	plainLines := make([]string, height)
	for i, line := range bgLines {
		plainLines[i] = ansi.Strip(line)
	}

	modalLines := strings.Split(modal, "\n")
	modalH := len(modalLines)
	modalW := lipgloss.Width(modal)
	startY := max(0, (height-modalH)/2)
	startX := max(0, (width-modalW)/2)

	result := make([]string, height)
	for y := 0; y < height; y++ {
		runes := []rune(plainLines[y])
		for len(runes) < width {
			runes = append(runes, ' ')
		}
		if len(runes) > width {
			runes = runes[:width]
		}

		if y >= startY && y < startY+modalH {
			mi := y - startY
			mLine := modalLines[mi]
			mLineW := lipgloss.Width(mLine)

			leftText := string(runes[:startX])
			rightStart := startX + mLineW
			rightText := ""
			if rightStart < len(runes) {
				rightText = string(runes[rightStart:])
			}

			result[y] = dimFg.Render(leftText) + mLine + dimFg.Render(rightText)
		} else {
			result[y] = dimFg.Render(string(runes))
		}
	}

	return strings.Join(result, "\n")
}

// overlayAt places content at a specific position over the background without
// dimming. Used for lightweight tooltips.
func overlayAt(bg string, content string, x, y, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}
	bgLines = bgLines[:height]

	contentLines := strings.Split(content, "\n")

	for i, cLine := range contentLines {
		row := y + i
		if row < 0 || row >= height {
			continue
		}

		bgRunes := []rune(ansi.Strip(bgLines[row]))
		for len(bgRunes) < width {
			bgRunes = append(bgRunes, ' ')
		}

		cLineW := lipgloss.Width(cLine)
		leftEnd := x
		if leftEnd > len(bgRunes) {
			leftEnd = len(bgRunes)
		}
		rightStart := x + cLineW
		if rightStart > len(bgRunes) {
			rightStart = len(bgRunes)
		}

		line := string(bgRunes[:leftEnd]) + cLine
		if rightStart < len(bgRunes) {
			line += string(bgRunes[rightStart:])
		}
		bgLines[row] = line
	}

	return strings.Join(bgLines, "\n")
}
