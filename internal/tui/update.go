package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andreasbaumgartner/gilo/internal/github"
	"github.com/andreasbaumgartner/gilo/internal/settings"
	"github.com/andreasbaumgartner/gilo/internal/tmux"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.detailInnerW()
		m.viewport.Height = m.detailInnerH()

	case tmuxTickMsg:
		return m, tea.Batch(fetchTmuxStatusCmd, tmuxTickCmd())

	case refreshTickMsg:
		m.refreshing = true
		return m, tea.Batch(fetchIssuesCmd(m.settings.GetIssuesMax()), refreshTickCmd())

	case tmuxStatusMsg:
		m.tmuxPanes = msg
		if m.loaded && len(m.issues) > 0 {
			m.updateViewport()
		}

	case tmuxWindowKilledMsg:
		if msg.ok {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchTmuxStatusCmd
		}
		m.modalStatus = fmt.Sprintf("Failed to close tmux window: %s", msg.windowName)

	case tmuxJumpMsg:
		if msg.ok {
			m.modal = modalNone
			m.modalStatus = ""
		} else {
			m.modalStatus = fmt.Sprintf("Failed to switch to tmux window: %s", msg.windowName)
		}

	case issuesLoadedMsg:
		m.issues = []github.Issue(msg)
		m.loaded = true
		m.refreshing = false
		filtered := m.filteredIssues()
		if m.cursor >= len(filtered) {
			m.cursor = max(0, len(filtered)-1)
		}
		m.clampListOffset()
		m.updateViewport()

	case errMsg:
		m.err = msg.err
		m.refreshing = false

	case browserOpenedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
		}

	case commentPostedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case issueCreatedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case issueDeletedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case issueClosedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case issueReopenedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case labelsLoadedMsg:
		m.repoLabels = msg.repoLabels
		m.labelSelected = make(map[string]bool)
		m.labelOriginal = make(map[string]bool)
		for k, v := range msg.issueLabels {
			m.labelSelected[k] = v
			m.labelOriginal[k] = v
		}
		m.modalStatus = ""

	case labelsUpdatedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case worktreeCreatedMsg:
		if msg.Err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			verb := "Created"
			if msg.Existed {
				verb = "Opened existing"
			}
			if msg.Session != "" {
				action := "Attach with: tmux attach -t " + msg.Session
				if msg.Reused {
					action = "Session already running — attach with:\ntmux attach -t " + msg.Session
				}
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch:  %s\nPath:    %s\nSession: %s\n\n%s", verb, msg.Branch, msg.Path, msg.Session, action)
			} else if msg.Reused {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nJumped to existing tmux window.", verb, msg.Branch, msg.Path)
			} else if msg.Split {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nOpened in tmux split pane.", verb, msg.Branch, msg.Path)
			} else {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nOpened in new tmux window.", verb, msg.Branch, msg.Path)
			}
		}

	case claudeTaskCreatedMsg:
		if msg.Err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			verb := "Created"
			if msg.Existed {
				verb = "Opened existing"
			}
			if msg.Reused {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nClaude already running in tmux window.", verb, msg.Branch, msg.Path)
			} else if msg.Split {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nClaude started in tmux split pane.", verb, msg.Branch, msg.Path)
			} else {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nClaude started in background tmux window.", verb, msg.Branch, msg.Path)
			}
		}

	case tea.KeyMsg:
		if m.modal != modalNone {
			return m.updateModal(msg)
		}
		return m.updateNormal(msg)
	}

	return m, nil
}

func (m model) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case modalHelp:
		if msg.String() == "esc" || msg.String() == "?" {
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalBrowser:
		if msg.String() == "esc" {
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalWorktree, modalClaudeTask:
		if msg.String() == "esc" {
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalPermissionWarning:
		switch msg.String() {
		case "y", "Y":
			m.settings.DangerouslySkipPermissions = true
			m.settings.PermissionWarningAcked = true
			settings.Save(m.settings)
			m.modal = modalNone
			m.modalStatus = ""
		case "n", "N", "esc":
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalDeleteConfirm:
		switch msg.String() {
		case "y", "Y":
			m.modalStatus = "Deleting issue..."
			num := m.modalIssue
			return m, deleteIssueCmd(num)
		case "n", "N", "esc":
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalKillWindowConfirm:
		switch msg.String() {
		case "y", "Y":
			windowName := m.modalKillWindowName
			m.modalStatus = "Closing tmux window..."
			return m, func() tea.Msg {
				ok := tmux.KillWindow(windowName)
				return tmuxWindowKilledMsg{windowName: windowName, ok: ok}
			}
		case "n", "N", "esc":
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalCloseConfirm:
		switch msg.String() {
		case "y", "Y":
			num := m.modalIssue
			if m.modalCloseAction == "close" {
				m.modalStatus = "Closing issue..."
				return m, closeIssueCmd(num)
			}
			m.modalStatus = "Reopening issue..."
			return m, reopenIssueCmd(num)
		case "n", "N", "esc":
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalComment:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			m.modalStatus = ""
			return m, nil
		case "ctrl+d":
			body := strings.TrimSpace(m.textarea.Value())
			if body == "" {
				return m, nil
			}
			m.modalStatus = "Posting comment..."
			num := m.modalIssue
			return m, func() tea.Msg {
				return commentPostedMsg{github.PostComment(num, body)}
			}
		}
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	case modalCreate:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			m.modalStatus = ""
			return m, nil
		case "tab":
			if m.createFocus == 0 {
				m.createFocus = 1
				m.textarea.Blur()
				m.textareaBody.Focus()
			} else {
				m.createFocus = 0
				m.textareaBody.Blur()
				m.textarea.Focus()
			}
			return m, nil
		case "ctrl+d":
			title := strings.TrimSpace(m.textarea.Value())
			if title == "" {
				return m, nil
			}
			body := strings.TrimSpace(m.textareaBody.Value())
			m.modalStatus = "Creating issue..."
			return m, func() tea.Msg {
				return issueCreatedMsg{github.CreateIssue(title, body)}
			}
		}
		var cmd tea.Cmd
		if m.createFocus == 0 {
			m.textarea, cmd = m.textarea.Update(msg)
		} else {
			m.textareaBody, cmd = m.textareaBody.Update(msg)
		}
		return m, cmd

	case modalLabel:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			m.modalStatus = ""
			return m, nil
		case "j", "down":
			if m.labelCursor < len(m.repoLabels)-1 {
				m.labelCursor++
			}
			return m, nil
		case "k", "up":
			if m.labelCursor > 0 {
				m.labelCursor--
			}
			return m, nil
		case " ", "enter":
			if len(m.repoLabels) > 0 {
				name := m.repoLabels[m.labelCursor].Name
				m.labelSelected[name] = !m.labelSelected[name]
			}
			return m, nil
		case "ctrl+d":
			var toAdd, toRemove []string
			for _, l := range m.repoLabels {
				was := m.labelOriginal[l.Name]
				now := m.labelSelected[l.Name]
				if now && !was {
					toAdd = append(toAdd, l.Name)
				} else if !now && was {
					toRemove = append(toRemove, l.Name)
				}
			}
			if len(toAdd) == 0 && len(toRemove) == 0 {
				m.modal = modalNone
				m.modalStatus = ""
				return m, nil
			}
			m.modalStatus = "Updating labels..."
			return m, applyLabelChangesCmd(m.modalIssue, toAdd, toRemove)
		}
		return m, nil
	}
	return m, nil
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.focus == focusList {
			m.focus = focusDetail
		} else {
			m.focus = focusList
		}

	case "j", "down":
		if m.focus == focusList {
			filtered := m.filteredIssues()
			if m.cursor < len(filtered)-1 {
				m.cursor++
				m.clampListOffset()
				m.updateViewport()
			}
		} else {
			m.viewport.ScrollDown(1)
		}

	case "k", "up":
		if m.focus == focusList {
			if m.cursor > 0 {
				m.cursor--
				m.clampListOffset()
				m.updateViewport()
			}
		} else {
			m.viewport.ScrollUp(1)
		}

	case "enter":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			// If a tmux session exists for this issue, jump to it.
			for _, p := range m.tmuxPanes {
				if p.IssueNum == issue.Number {
					windowName := p.WindowName
					m.modal = modalWorktree
					m.modalIssue = issue.Number
					m.modalStatus = "Jumping to tmux window..."
					return m, func() tea.Msg {
						ok := tmux.SelectWindow(windowName)
						return tmuxJumpMsg{windowName: windowName, ok: ok}
					}
				}
			}
			// No tmux session — fall back to opening in browser.
			m.modal = modalBrowser
			m.modalIssue = issue.Number
			m.modalStatus = ""
			num := m.modalIssue
			return m, func() tea.Msg {
				return browserOpenedMsg{github.OpenInBrowser(num)}
			}
		}

	case "o":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			m.modal = modalBrowser
			m.modalIssue = filtered[m.cursor].Number
			m.modalStatus = ""
			num := m.modalIssue
			return m, func() tea.Msg {
				return browserOpenedMsg{github.OpenInBrowser(num)}
			}
		}

	case "w", "W":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalWorktree
			m.modalIssue = issue.Number
			m.modalStatus = ""
			return m, createWorktreeCmd(issue, msg.String() == "W")
		}

	case "s", "S":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalClaudeTask
			m.modalIssue = issue.Number
			m.modalStatus = ""
			return m, createClaudeTaskCmd(issue, msg.String() == "S", m.settings.DangerouslySkipPermissions)
		}

	case "p":
		if m.settings.DangerouslySkipPermissions {
			m.settings.DangerouslySkipPermissions = false
			settings.Save(m.settings)
		} else if m.settings.PermissionWarningAcked {
			m.settings.DangerouslySkipPermissions = true
			settings.Save(m.settings)
		} else {
			m.modal = modalPermissionWarning
			m.modalStatus = ""
		}
		return m, nil

	case "c":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			m.modal = modalComment
			m.modalIssue = filtered[m.cursor].Number
			m.modalStatus = ""
			ta := textarea.New()
			ta.Placeholder = "Write your comment..."
			ta.Focus()
			ta.SetWidth(m.textareaW())
			ta.SetHeight(8)
			m.textarea = ta
		}

	case "n":
		m.modal = modalCreate
		m.modalStatus = ""
		m.createFocus = 0

		ta := textarea.New()
		ta.Placeholder = "Issue title..."
		ta.Focus()
		ta.SetWidth(m.textareaW())
		ta.SetHeight(1)
		ta.CharLimit = 256
		m.textarea = ta

		tb := textarea.New()
		tb.Placeholder = "Describe the issue..."
		tb.Blur()
		tb.SetWidth(m.textareaW())
		tb.SetHeight(6)
		m.textareaBody = tb

	case "l":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalLabel
			m.modalIssue = issue.Number
			m.modalStatus = "Loading labels..."
			m.labelCursor = 0
			return m, fetchLabelsCmd(issue)
		}

	case "x":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalCloseConfirm
			m.modalIssue = issue.Number
			m.modalStatus = ""
			if issue.State == "OPEN" {
				m.modalCloseAction = "close"
			} else {
				m.modalCloseAction = "reopen"
			}
		}
		return m, nil

	case "d":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			m.modal = modalDeleteConfirm
			m.modalIssue = filtered[m.cursor].Number
			m.modalStatus = ""
		}
		return m, nil

	case "K":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			for _, p := range m.tmuxPanes {
				if p.IssueNum == issue.Number {
					m.modal = modalKillWindowConfirm
					m.modalIssue = issue.Number
					m.modalKillWindowName = p.WindowName
					m.modalStatus = ""
					return m, nil
				}
			}
		}
		return m, nil

	case "?":
		m.modal = modalHelp
		m.modalStatus = ""
		return m, nil

	case "g":
		if !m.refreshing {
			m.refreshing = true
			return m, fetchIssuesCmd(m.settings.GetIssuesMax())
		}

	case "f":
		switch m.stateFilter {
		case "OPEN":
			m.stateFilter = "CLOSED"
		case "CLOSED":
			m.stateFilter = ""
		default:
			m.stateFilter = "OPEN"
		}
		filtered := m.filteredIssues()
		if m.cursor >= len(filtered) {
			m.cursor = max(0, len(filtered)-1)
		}
		m.listOffset = 0
		m.clampListOffset()
		m.updateViewport()

	case "a":
		// Cycle to next sort column (descending by default)
		m.sortCol = (m.sortCol + 1) % sortColumn(len(sortColumnNames))
		m.sortAsc = false
		m.cursor = 0
		m.listOffset = 0
		m.clampListOffset()
		m.updateViewport()

	case "A":
		// Toggle sort direction for current column
		m.sortAsc = !m.sortAsc
		m.cursor = 0
		m.listOffset = 0
		m.clampListOffset()
		m.updateViewport()

	case "t":
		current := m.settings.ColorScheme
		if current == "" {
			current = "default"
		}
		next := colorSchemeOrder[0]
		for i, name := range colorSchemeOrder {
			if name == current {
				next = colorSchemeOrder[(i+1)%len(colorSchemeOrder)]
				break
			}
		}
		applyColorScheme(colorSchemes[next])
		m.settings.ColorScheme = next
		settings.Save(m.settings)
		m.updateViewport()

	case "m":
		current := m.settings.GetIssuesMax()
		opts := settings.IssuesMaxOptions
		next := opts[0]
		for i, v := range opts {
			if v == current {
				next = opts[(i+1)%len(opts)]
				break
			}
		}
		m.settings.IssuesMax = next
		settings.Save(m.settings)
		m.refreshing = true
		return m, fetchIssuesCmd(next)
	}

	return m, nil
}

func (m *model) clampListOffset() {
	innerH := m.mainH() - 2
	// 2 rows for header + blank line
	visibleRows := innerH - 2

	// Account for scroll indicator rows
	if m.listOffset > 0 {
		visibleRows-- // up indicator
	}
	filtered := m.filteredIssues()
	if m.listOffset+visibleRows < len(filtered) {
		visibleRows-- // down indicator
	}
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Scroll down if cursor is below visible area
	if m.cursor >= m.listOffset+visibleRows {
		m.listOffset = m.cursor - visibleRows + 1
	}
	// Scroll up if cursor is above visible area
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	// Clamp offset
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

func (m *model) updateViewport() {
	m.viewport.Width = m.detailInnerW()
	m.viewport.Height = m.detailInnerH()
	m.viewport.SetContent(m.renderDetailContent())
	m.viewport.GotoTop()
}
