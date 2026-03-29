package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/andreasbaumgartner/gilo/internal/tmux"
)

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf(
			"\n  Error: %v\n\n  Make sure 'gh' is installed and authenticated.\n  Press q to quit.\n",
			m.err,
		)
	}

	left := m.renderList()
	right := m.renderDetail()
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	screen := body + "\n" + m.renderStatusBar()

	if m.modal != modalNone {
		modal := m.renderModal()
		screen = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(activeScheme.OverlayBg),
		)

		lines := strings.Split(screen, "\n")
		for len(lines) < m.height {
			lines = append(lines, "")
		}
		lines[m.height-1] = m.renderStatusBar()
		screen = strings.Join(lines, "\n")
	}

	return screen
}

// Modal

func (m model) renderModal() string {
	switch m.modal {
	case modalBrowser:
		content := fmt.Sprintf("Opening issue #%d in browser...", m.modalIssue)
		if m.modalStatus != "" {
			content = m.modalStatus
		}
		hint := dimStyle.Render("esc dismiss")
		return modalStyle.Render(content + "\n\n" + hint)

	case modalWorktree:
		title := titleStyle.Render(fmt.Sprintf("Worktree for #%d", m.modalIssue))
		hint := dimStyle.Render("esc dismiss")
		content := "Setting up worktree..."
		if m.modalStatus != "" {
			content = m.modalStatus
		}
		return modalStyle.Render(strings.Join([]string{title, "", content, "", hint}, "\n"))

	case modalClaudeTask:
		title := titleStyle.Render(fmt.Sprintf("Claude Task for #%d", m.modalIssue))
		hint := dimStyle.Render("esc dismiss")
		content := "Setting up worktree and starting Claude..."
		if m.modalStatus != "" {
			content = m.modalStatus
		}
		if m.settings.DangerouslySkipPermissions {
			content += "\n\n" + lipgloss.NewStyle().Foreground(colorYellow).Render("⚡ Running with --dangerously-skip-permissions")
		}
		return modalStyle.Render(strings.Join([]string{title, "", content, "", hint}, "\n"))

	case modalPermissionWarning:
		title := titleStyle.Render("⚠ Enable All Permissions")
		warning := lipgloss.NewStyle().Foreground(colorYellow).Render(
			"WARNING: This will run Claude with --dangerously-skip-permissions.\n\n" +
				"Claude will be able to execute any tool (shell commands, file\n" +
				"writes, etc.) without asking for your confirmation.\n\n" +
				"Only enable this if you trust the environment and understand\n" +
				"the risks.")
		hint := dimStyle.Render("y confirm  │  n/esc cancel")
		return modalStyle.Render(strings.Join([]string{title, "", warning, "", hint}, "\n"))

	case modalCloseConfirm:
		action := "Close"
		prompt := "Are you sure you want to close this issue?"
		if m.modalCloseAction == "reopen" {
			action = "Reopen"
			prompt = "Are you sure you want to reopen this issue?"
		}
		title := titleStyle.Render(fmt.Sprintf("%s Issue #%d", action, m.modalIssue))
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else {
			body = prompt
		}
		hint := dimStyle.Render("y confirm  │  n/esc cancel")
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))

	case modalDeleteConfirm:
		title := titleStyle.Render(fmt.Sprintf("Delete Issue #%d", m.modalIssue))
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else {
			body = "Are you sure you want to delete this issue?\nThis action cannot be undone."
		}
		hint := dimStyle.Render("y confirm  │  n/esc cancel")
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))

	case modalComment:
		title := titleStyle.Render(fmt.Sprintf("Comment on #%d", m.modalIssue))
		hint := dimStyle.Render("ctrl+d submit  │  esc cancel")
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else {
			body = m.textarea.View()
		}
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))

	case modalCreate:
		title := titleStyle.Render("Create New Issue")
		hint := dimStyle.Render("tab switch field  │  ctrl+d submit  │  esc cancel")
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else {
			body = strings.Join([]string{
				dimStyle.Render("Title:"),
				m.textarea.View(),
				"",
				dimStyle.Render("Body:"),
				m.textareaBody.View(),
			}, "\n")
		}
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))

	case modalLabel:
		title := titleStyle.Render(fmt.Sprintf("Labels for #%d", m.modalIssue))
		hint := dimStyle.Render("j/k navigate  │  space toggle  │  ctrl+d submit  │  esc cancel")
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else if len(m.repoLabels) == 0 {
			body = dimStyle.Render("No labels found in this repository.")
		} else {
			var rows []string
			for i, l := range m.repoLabels {
				check := "[ ]"
				if m.labelSelected[l.Name] {
					check = "[x]"
				}
				labelColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#" + l.Color))
				line := fmt.Sprintf(" %s %s", check, labelColor.Render(l.Name))
				if i == m.labelCursor {
					line = selectedStyle.Render(fmt.Sprintf(" %s %s", check, l.Name))
				}
				rows = append(rows, line)
			}
			body = strings.Join(rows, "\n")
		}
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))
	case modalHelp:
		title := titleStyle.Render("Keybindings")
		sections := []string{
			title, "",
			dimStyle.Render("── Navigation ──"),
			"  ↑/k       Move up",
			"  ↓/j       Move down",
			"  tab       Switch panel",
			"  f         Cycle filter (Open/Closed/All)",
			"",
			dimStyle.Render("── Actions ──"),
			"  enter     Jump to tmux window / open in browser",
			"  o         Open issue in browser",
			"  c         Comment on issue",
			"  n         Create new issue",
			"  x         Close/reopen issue",
			"  d         Delete issue",
			"  l         Manage labels",
			"",
			dimStyle.Render("── Worktree & Claude ──"),
			"  w         Worktree in new tmux window",
			"  W         Worktree in tmux split",
			"  s         Claude session in new window",
			"  S         Claude session in split",
			"",
			dimStyle.Render("── Settings ──"),
			"  p         Toggle skip-permissions",
			"  r         Toggle allow-root",
			"  t         Cycle color scheme",
			"",
			dimStyle.Render("── General ──"),
			"  ?         Show this help",
			"  q         Quit",
			"",
			dimStyle.Render("esc/?  close"),
		}
		return modalStyle.Width(46).Render(strings.Join(sections, "\n"))
	}
	return ""
}

// List panel

func (m model) renderList() string {
	active := m.focus == focusList
	innerW := m.listInnerW()
	innerH := m.mainH() - 2

	filterLabel := "Open"
	if m.stateFilter == "CLOSED" {
		filterLabel = "Closed"
	} else if m.stateFilter == "" {
		filterLabel = "All"
	}

	refreshIndicator := ""
	if m.refreshing {
		refreshIndicator = " " + yellowStyle.Render("⟳")
	}

	header := titleStyle.Render("Issues") + " " + dimStyle.Render("["+filterLabel+"]") + refreshIndicator
	if !active {
		header = dimStyle.Render("Issues") + " " + dimStyle.Render("["+filterLabel+"]") + refreshIndicator
	}

	var rows []string
	rows = append(rows, header, "")

	filtered := m.filteredIssues()
	if !m.loaded {
		rows = append(rows, "  "+dimStyle.Render("Loading..."))
	} else if len(filtered) == 0 {
		rows = append(rows, "  "+dimStyle.Render("No "+strings.ToLower(filterLabel)+" issues found."))
	} else {
		visibleRows := innerH - len(rows)
		start := m.listOffset

		// Reserve space for scroll indicators
		if start > 0 {
			visibleRows--
		}
		if start+visibleRows < len(filtered) {
			visibleRows--
		}
		if visibleRows < 1 {
			visibleRows = 1
		}

		end := start + visibleRows
		if end > len(filtered) {
			end = len(filtered)
		}

		if start > 0 {
			rows = append(rows, dimStyle.Render(fmt.Sprintf("  ↑ %d more issue(s)", start)))
		}

		for idx := start; idx < end; idx++ {
			i := idx
			issue := filtered[idx]

			stateBadge := openBadge.Render("OPEN")
			if issue.State == "CLOSED" {
				stateBadge = closedBadge.Render("CLOSED")
			}

			statusBadge := ""
			for _, p := range m.tmuxPanes {
				if p.IssueNum == issue.Number {
					switch p.Status {
					case tmux.StatusReview:
						statusBadge = reviewBadge.Render("REVIEW")
					case tmux.StatusQuestion:
						statusBadge = questionBadge.Render("QUESTION")
					default:
						statusBadge = workingBadge.Render("WORKING")
					}
					break
				}
			}

			// Fixed-width columns for consistent alignment
			const statusColWidth = 10
			statusCol := padRight(" "+statusBadge, statusColWidth)

			pad := ""
			if issue.State != "CLOSED" {
				pad = "  "
			}

			num := dimStyle.Render(fmt.Sprintf("#%-4d", issue.Number))
			title := truncate(issue.Title, innerW-15-statusColWidth)
			line := stateBadge + pad + statusCol + " " + num + " " + title

			if i == m.cursor {
				if active {
					rest := fmt.Sprintf("#%-4d %s", issue.Number, truncate(issue.Title, innerW-15-statusColWidth))
					line = stateBadge + pad + statusCol + " " + selectedStyle.Render(padRight(rest, innerW-9-statusColWidth))
				} else {
					line = lipgloss.NewStyle().
						Foreground(activeScheme.UnfocusedSelected).
						Render(line)
				}
			}
			rows = append(rows, line)
		}

		if end < len(filtered) {
			remaining := len(filtered) - end
			rows = append(rows, dimStyle.Render(fmt.Sprintf("  ↓ %d more issue(s)", remaining)))
		}
	}

	content := strings.Join(rows, "\n")
	return panelBorder(active).
		Width(m.listW()-2).
		Height(m.mainH()-2).
		Padding(0, 1).
		Render(content)
}

// Detail panel

func (m model) renderDetail() string {
	active := m.focus == focusDetail
	return panelBorder(active).
		Width(m.detailW()-2).
		Height(m.mainH()-2).
		Padding(0, 1).
		Render(m.viewport.View())
}

func (m model) renderDetailContent() string {
	filtered := m.filteredIssues()
	if !m.loaded || len(filtered) == 0 {
		return dimStyle.Render("No issue selected.")
	}

	issue := filtered[m.cursor]
	w := m.detailInnerW()

	stateColor := greenStyle
	if issue.State == "CLOSED" {
		stateColor = redStyle
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("#%d  %s", issue.Number, issue.Title)))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(w, 60)))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("State:   ") + stateColor.Render(issue.State) + "\n")
	b.WriteString(dimStyle.Render("Author:  ") + issue.Author.Login + "\n")
	b.WriteString(dimStyle.Render("Created: ") + issue.CreatedAt[:10] + "\n")

	if len(issue.Labels) > 0 {
		names := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			names[i] = l.Name
		}
		b.WriteString(dimStyle.Render("Labels:  ") + yellowStyle.Render(strings.Join(names, ", ")) + "\n")
	}

	for _, p := range m.tmuxPanes {
		if p.IssueNum == issue.Number {
			b.WriteString("\n")
			b.WriteString(dimStyle.Render("── Tmux ") + dimStyle.Render(strings.Repeat("─", max(0, w-10))) + "\n\n")
			var statusLabel string
			switch p.Status {
			case tmux.StatusReview:
				statusLabel = reviewBadge.Render("REVIEW")
			case tmux.StatusQuestion:
				statusLabel = questionBadge.Render("QUESTION")
			default:
				statusLabel = workingBadge.Render("WORKING")
			}
			b.WriteString("  " + yellowStyle.Render("Status: ") + statusLabel + "\n")
			b.WriteString("  " + yellowStyle.Render("Window: ") + p.WindowName + "\n")
			if p.LastLine != "" {
				b.WriteString("  " + yellowStyle.Render("Output: ") + truncate(p.LastLine, w-12) + "\n")
			}
			break
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("── Description ") + dimStyle.Render(strings.Repeat("─", max(0, w-17))) + "\n\n")
	if issue.Body == "" {
		b.WriteString(dimStyle.Render("  (no description)\n"))
	} else {
		for _, line := range strings.Split(issue.Body, "\n") {
			b.WriteString("  " + line + "\n")
		}
	}

	if len(issue.Comments) > 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("── Comments (%d) ", len(issue.Comments))) +
			dimStyle.Render(strings.Repeat("─", max(0, w-20))) + "\n")

		for _, c := range issue.Comments {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(c.Author.Login))
			b.WriteString("  " + dimStyle.Render(c.CreatedAt[:10]) + "\n")
			for _, line := range strings.Split(c.Body, "\n") {
				b.WriteString("  " + line + "\n")
			}
		}
	}

	return b.String()
}

// Status bar

func (m model) renderStatusBar() string {
	var keys []string
	switch m.modal {
	case modalComment, modalCreate:
		keys = []string{"tab switch field", "ctrl+d submit", "esc cancel"}
	case modalLabel:
		keys = []string{"j/k navigate", "space toggle", "ctrl+d submit", "esc cancel"}
	case modalHelp:
		keys = []string{"esc/? close"}
	case modalDeleteConfirm, modalCloseConfirm:
		keys = []string{"y confirm", "n/esc cancel"}
	case modalBrowser, modalWorktree, modalClaudeTask:
		keys = []string{"esc dismiss"}
	case modalPermissionWarning:
		keys = []string{"y confirm", "n/esc cancel"}
	default:
		permLabel := "p permissions:off"
		if m.settings.DangerouslySkipPermissions {
			permLabel = "p permissions:ON"
		}
		keys = []string{
			"↑↓/jk navigate",
			"tab switch panel",
			"f filter state",
			"g refresh",
			"enter jump/open",
			"o browser",
			"c comment",
			"w/W worktree/split",
			"s/S claude/split",
			permLabel,
			"t theme:" + activeScheme.Name,
			"l labels",
			"x close/reopen",
			"d delete",
			"n new issue",
			"? help",
			"q quit",
		}
	}
	if m.modal == modalNone && len(m.tmuxPanes) > 0 {
		keys = append([]string{fmt.Sprintf("%d tmux ⟳", len(m.tmuxPanes))}, keys...)
	}
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = keybindStyle.Render(k)
	}
	bar := strings.Join(parts, " ")
	barW := lipgloss.Width(bar)
	if barW < m.width {
		bar += strings.Repeat(" ", m.width-barW)
	}
	return bar
}
