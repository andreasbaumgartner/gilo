package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andreasbaumgartner/gilo/internal/github"
	"github.com/andreasbaumgartner/gilo/internal/tmux"
)

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
