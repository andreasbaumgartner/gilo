package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andreasbaumgartner/gilo/internal/github"
	"github.com/andreasbaumgartner/gilo/internal/tmux"
)

func TestPrefixKeyFilterSort(t *testing.T) {
	baseModel := func() model {
		return model{
			width:       100,
			height:      50,
			loaded:      true,
			stateFilter: "OPEN",
			issues: []github.Issue{
				{Number: 1, Title: "Alpha", State: "OPEN", CreatedAt: "2025-01-01T00:00:00Z"},
				{Number: 2, Title: "Beta", State: "OPEN", CreatedAt: "2025-01-02T00:00:00Z"},
				{Number: 3, Title: "Gamma", State: "CLOSED", CreatedAt: "2025-01-03T00:00:00Z"},
			},
		}
	}

	t.Run("f sets pending key", func(t *testing.T) {
		m := baseModel()
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)
		if rm.pendingKey != "f" {
			t.Errorf("expected pendingKey = %q, got %q", "f", rm.pendingKey)
		}
	})

	t.Run("ff cycles filter", func(t *testing.T) {
		m := baseModel()
		m.pendingKey = "f"
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)
		if rm.stateFilter != "CLOSED" {
			t.Errorf("expected stateFilter = %q, got %q", "CLOSED", rm.stateFilter)
		}
		if rm.pendingKey != "" {
			t.Errorf("expected pendingKey cleared, got %q", rm.pendingKey)
		}
	})

	t.Run("fs cycles sort column", func(t *testing.T) {
		m := baseModel()
		m.pendingKey = "f"
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)
		if rm.sortCol != sortByTitle {
			t.Errorf("expected sortCol = sortByTitle, got %d", rm.sortCol)
		}
		if rm.pendingKey != "" {
			t.Errorf("expected pendingKey cleared, got %q", rm.pendingKey)
		}
	})

	t.Run("fd toggles sort direction", func(t *testing.T) {
		m := baseModel()
		m.pendingKey = "f"
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)
		if !rm.sortAsc {
			t.Error("expected sortAsc = true after toggle")
		}
		if rm.pendingKey != "" {
			t.Errorf("expected pendingKey cleared, got %q", rm.pendingKey)
		}
	})

	t.Run("f then unknown key cancels pending", func(t *testing.T) {
		m := baseModel()
		m.pendingKey = "f"
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)
		if rm.pendingKey != "" {
			t.Errorf("expected pendingKey cleared, got %q", rm.pendingKey)
		}
		// Filter should be unchanged
		if rm.stateFilter != "OPEN" {
			t.Errorf("expected stateFilter unchanged, got %q", rm.stateFilter)
		}
	})
}

func TestAdditionalContextToggle(t *testing.T) {
	t.Run("i toggles additional context on", func(t *testing.T) {
		m := model{
			width:  100,
			height: 50,
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)

		if !rm.settings.AdditionalContext {
			t.Error("expected AdditionalContext to be true after toggle")
		}
	})

	t.Run("i toggles additional context off", func(t *testing.T) {
		m := model{
			width:  100,
			height: 50,
		}
		m.settings.AdditionalContext = true

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)

		if rm.settings.AdditionalContext {
			t.Error("expected AdditionalContext to be false after toggle")
		}
	})
}

func TestKillWindowKeybind(t *testing.T) {
	t.Run("K opens confirmation when tmux pane exists", func(t *testing.T) {
		m := model{
			width:       100,
			height:      50,
			loaded:      true,
			stateFilter: "OPEN",
			issues: []github.Issue{
				{Number: 42, Title: "Test issue", State: "OPEN"},
			},
			tmuxPanes: []tmux.Pane{
				{WindowName: "issue-42-test-issue", IssueNum: 42},
			},
			cursor: 0,
			modal:  modalNone,
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)

		if rm.modal != modalKillWindowConfirm {
			t.Errorf("expected modal = modalKillWindowConfirm, got %d", rm.modal)
		}
		if rm.modalIssue != 42 {
			t.Errorf("expected modalIssue = 42, got %d", rm.modalIssue)
		}
		if rm.modalKillWindowName != "issue-42-test-issue" {
			t.Errorf("expected modalKillWindowName = %q, got %q", "issue-42-test-issue", rm.modalKillWindowName)
		}
	})

	t.Run("K does nothing without tmux pane", func(t *testing.T) {
		m := model{
			width:       100,
			height:      50,
			loaded:      true,
			stateFilter: "OPEN",
			issues: []github.Issue{
				{Number: 42, Title: "Test issue", State: "OPEN"},
			},
			tmuxPanes: []tmux.Pane{},
			cursor:    0,
			modal:     modalNone,
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)

		if rm.modal != modalNone {
			t.Errorf("expected modal = modalNone, got %d", rm.modal)
		}
	})

	t.Run("K does nothing with no issues", func(t *testing.T) {
		m := model{
			width:       100,
			height:      50,
			loaded:      true,
			stateFilter: "OPEN",
			issues:      []github.Issue{},
			cursor:      0,
			modal:       modalNone,
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}}
		result, _ := m.updateNormal(msg)
		rm := result.(model)

		if rm.modal != modalNone {
			t.Errorf("expected modal = modalNone, got %d", rm.modal)
		}
	})
}

func TestKillWindowModalDismiss(t *testing.T) {
	t.Run("esc cancels kill window modal", func(t *testing.T) {
		m := model{
			modal:               modalKillWindowConfirm,
			modalIssue:          42,
			modalKillWindowName: "issue-42-test",
		}

		msg := tea.KeyMsg{Type: tea.KeyEscape}
		result, _ := m.updateModal(msg)
		rm := result.(model)

		if rm.modal != modalNone {
			t.Errorf("expected modal = modalNone after esc, got %d", rm.modal)
		}
	})

	t.Run("n cancels kill window modal", func(t *testing.T) {
		m := model{
			modal:               modalKillWindowConfirm,
			modalIssue:          42,
			modalKillWindowName: "issue-42-test",
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := m.updateModal(msg)
		rm := result.(model)

		if rm.modal != modalNone {
			t.Errorf("expected modal = modalNone after n, got %d", rm.modal)
		}
	})

	t.Run("y confirms kill and returns command", func(t *testing.T) {
		m := model{
			modal:               modalKillWindowConfirm,
			modalIssue:          42,
			modalKillWindowName: "issue-42-test",
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
		result, cmd := m.updateModal(msg)
		rm := result.(model)

		if rm.modalStatus != "Closing tmux window..." {
			t.Errorf("expected modalStatus = %q, got %q", "Closing tmux window...", rm.modalStatus)
		}
		if cmd == nil {
			t.Error("expected a command to be returned after confirming kill")
		}
	})
}

func TestTmuxWindowKilledMsg(t *testing.T) {
	t.Run("successful kill dismisses modal", func(t *testing.T) {
		m := model{
			modal:               modalKillWindowConfirm,
			modalIssue:          42,
			modalKillWindowName: "issue-42-test",
			modalStatus:         "Closing tmux window...",
		}

		result, _ := m.Update(tmuxWindowKilledMsg{windowName: "issue-42-test", ok: true})
		rm := result.(model)

		if rm.modal != modalNone {
			t.Errorf("expected modal = modalNone after successful kill, got %d", rm.modal)
		}
		if rm.modalStatus != "" {
			t.Errorf("expected empty modalStatus, got %q", rm.modalStatus)
		}
	})

	t.Run("failed kill shows error", func(t *testing.T) {
		m := model{
			modal:               modalKillWindowConfirm,
			modalIssue:          42,
			modalKillWindowName: "issue-42-test",
			modalStatus:         "Closing tmux window...",
		}

		result, _ := m.Update(tmuxWindowKilledMsg{windowName: "issue-42-test", ok: false})
		rm := result.(model)

		if rm.modalStatus == "" {
			t.Error("expected non-empty modalStatus after failed kill")
		}
	})
}
