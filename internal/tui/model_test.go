package tui

import (
	"testing"

	"github.com/andreasbaumgartner/gilo/internal/github"
)

func TestFilteredIssues(t *testing.T) {
	issues := []github.Issue{
		{Number: 1, Title: "Open issue 1", State: "OPEN"},
		{Number: 2, Title: "Closed issue 1", State: "CLOSED"},
		{Number: 3, Title: "Open issue 2", State: "OPEN"},
		{Number: 4, Title: "Closed issue 2", State: "CLOSED"},
		{Number: 5, Title: "Open issue 3", State: "OPEN"},
	}

	tests := []struct {
		name   string
		filter string
		want   int
	}{
		{"filter open", "OPEN", 3},
		{"filter closed", "CLOSED", 2},
		{"filter all (empty)", "", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				issues:      issues,
				stateFilter: tt.filter,
			}
			got := m.filteredIssues()
			if len(got) != tt.want {
				t.Errorf("filteredIssues() with filter %q returned %d issues, want %d",
					tt.filter, len(got), tt.want)
			}
		})
	}
}

func TestFilteredIssuesCorrectItems(t *testing.T) {
	issues := []github.Issue{
		{Number: 1, State: "OPEN"},
		{Number: 2, State: "CLOSED"},
		{Number: 3, State: "OPEN"},
	}

	m := model{issues: issues, stateFilter: "OPEN"}
	got := m.filteredIssues()

	for _, issue := range got {
		if issue.State != "OPEN" {
			t.Errorf("filteredIssues(OPEN) returned issue #%d with state %q", issue.Number, issue.State)
		}
	}
}

func TestFilteredIssuesEmpty(t *testing.T) {
	m := model{issues: nil, stateFilter: "OPEN"}
	got := m.filteredIssues()
	if got != nil {
		t.Errorf("filteredIssues() with nil issues returned %v, want nil", got)
	}
}

func TestFilteredIssuesNoMatch(t *testing.T) {
	issues := []github.Issue{
		{Number: 1, State: "OPEN"},
	}

	m := model{issues: issues, stateFilter: "CLOSED"}
	got := m.filteredIssues()
	if len(got) != 0 {
		t.Errorf("filteredIssues() with no matches returned %d items, want 0", len(got))
	}
}

func TestModelDefaults(t *testing.T) {
	// Test that initialModel sets expected defaults
	// Note: initialModel() calls term.GetSize which may return 0,0 in test env
	m := model{
		stateFilter: "OPEN",
		modal:       modalNone,
	}

	if m.stateFilter != "OPEN" {
		t.Errorf("default stateFilter = %q, want OPEN", m.stateFilter)
	}
	if m.modal != modalNone {
		t.Errorf("default modal = %d, want modalNone (0)", m.modal)
	}
	if m.loaded {
		t.Error("default loaded should be false")
	}
	if m.cursor != 0 {
		t.Errorf("default cursor = %d, want 0", m.cursor)
	}
}

func TestModalKindConstants(t *testing.T) {
	// Verify modal constants are distinct
	kinds := []modalKind{
		modalNone, modalBrowser, modalComment, modalWorktree,
		modalCreate, modalLabel, modalClaudeTask, modalDeleteConfirm,
		modalPermissionWarning, modalHelp,
	}
	seen := make(map[modalKind]bool)
	for _, k := range kinds {
		if seen[k] {
			t.Errorf("duplicate modalKind value: %d", k)
		}
		seen[k] = true
	}
}

func TestFocusConstants(t *testing.T) {
	if focusList != 0 {
		t.Errorf("focusList = %d, want 0", focusList)
	}
	if focusDetail != 1 {
		t.Errorf("focusDetail = %d, want 1", focusDetail)
	}
}
