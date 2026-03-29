package tui

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/andreasbaumgartner/gilo/internal/github"
	"github.com/andreasbaumgartner/gilo/internal/settings"
	"github.com/andreasbaumgartner/gilo/internal/tmux"
)

// Messages

type issuesLoadedMsg []github.Issue
type errMsg struct{ err error }
type browserOpenedMsg struct{ err error }
type commentPostedMsg struct{ err error }
type issueCreatedMsg struct{ err error }
type issueDeletedMsg struct{ err error }
type labelsLoadedMsg struct {
	repoLabels  []github.RepoLabel
	issueLabels map[string]bool
}
type labelsUpdatedMsg struct{ err error }
type worktreeCreatedMsg struct {
	tmux.WorktreeResult
}
type claudeTaskCreatedMsg struct {
	tmux.ClaudeResult
}
type tmuxStatusMsg []tmux.Pane
type tmuxTickMsg struct{}
type refreshTickMsg struct{}

// Focus

type focus int

const (
	focusList focus = iota
	focusDetail
)

// Modal kinds

type modalKind int

const (
	modalNone modalKind = iota
	modalBrowser
	modalComment
	modalWorktree
	modalCreate
	modalLabel
	modalClaudeTask
	modalDeleteConfirm
	modalPermissionWarning
	modalHelp
)

// Model

type model struct {
	width  int
	height int

	focus  focus
	issues []github.Issue
	cursor int
	loaded bool
	err    error

	viewport viewport.Model

	modal       modalKind
	modalIssue  int
	modalStatus string
	textarea    textarea.Model

	textareaBody textarea.Model
	createFocus  int

	repoLabels    []github.RepoLabel
	labelSelected map[string]bool
	labelOriginal map[string]bool
	labelCursor   int

	tmuxPanes []tmux.Pane

	refreshing bool

	stateFilter string

	settings settings.Settings
}

func initialModel() model {
	w, h, _ := term.GetSize(0)
	vp := viewport.New(0, 0)
	s := settings.Load()
	if cs, ok := colorSchemes[s.ColorScheme]; ok {
		applyColorScheme(cs)
	}
	return model{
		width:       w,
		height:      h,
		viewport:    vp,
		modal:       modalNone,
		stateFilter: "OPEN",
		settings:    s,
	}
}

func (m model) filteredIssues() []github.Issue {
	if m.stateFilter == "" {
		return m.issues
	}
	var out []github.Issue
	for _, issue := range m.issues {
		if issue.State == m.stateFilter {
			out = append(out, issue)
		}
	}
	return out
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchIssuesCmd, fetchTmuxStatusCmd, tmuxTickCmd(), refreshTickCmd())
}

// Run starts the TUI application.
func Run() error {
	if err := exec.Command("git", "rev-parse", "--show-toplevel").Run(); err != nil {
		return fmt.Errorf("not a git repository. Please run gilo from inside a git repo")
	}

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
