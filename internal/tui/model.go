package tui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/andreasbaumgartner/gilo/internal/github"
	"github.com/andreasbaumgartner/gilo/internal/settings"
	"github.com/andreasbaumgartner/gilo/internal/tmux"
)

// Sort columns
type sortColumn int

const (
	sortByNumber sortColumn = iota
	sortByTitle
	sortByState
	sortByCreated
)

var sortColumnNames = []string{"Number", "Title", "State", "Created"}

// Messages

type issuesLoadedMsg []github.Issue
type errMsg struct{ err error }
type browserOpenedMsg struct{ err error }
type commentPostedMsg struct{ err error }
type issueCreatedMsg struct{ err error }
type issueDeletedMsg struct{ err error }
type issueClosedMsg struct{ err error }
type issueReopenedMsg struct{ err error }
type issueMergedMsg struct{ err error }
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
type tmuxJumpMsg struct {
	windowName string
	ok         bool
}
type tmuxWindowKilledMsg struct {
	windowName string
	ok         bool
}
type tmuxStatusMsg []tmux.Pane
type tmuxTickMsg struct{}
type refreshTickMsg struct{}
type linkedPRsMsg struct {
	prs map[int]github.PR
	err error
}

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
	modalCloseConfirm
	modalPermissionWarning
	modalKillWindowConfirm
	modalMerge
	modalMergeConfirm
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

	modal            modalKind
	modalIssue       int
	modalStatus      string
	modalCloseAction    string // "close" or "reopen"
	modalKillWindowName string // tmux window name to kill
	textarea            textarea.Model

	textareaBody textarea.Model
	createFocus  int

	mergeTarget int // issue number to merge into
	mergeCursor int // cursor in the merge target list

	repoLabels    []github.RepoLabel
	labelSelected map[string]bool
	labelOriginal map[string]bool
	labelCursor   int

	tmuxPanes []tmux.Pane
	linkedPRs map[int]github.PR

	refreshing bool

	listOffset int

	stateFilter string

	sortCol sortColumn
	sortAsc bool

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
	var out []github.Issue
	if m.stateFilter == "" {
		out = make([]github.Issue, len(m.issues))
		copy(out, m.issues)
	} else {
		for _, issue := range m.issues {
			if issue.State == m.stateFilter {
				out = append(out, issue)
			}
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		var less bool
		switch m.sortCol {
		case sortByTitle:
			less = strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
		case sortByState:
			less = out[i].State < out[j].State
		case sortByCreated:
			less = out[i].CreatedAt < out[j].CreatedAt
		default: // sortByNumber
			less = out[i].Number < out[j].Number
		}
		if m.sortAsc {
			return less
		}
		return !less
	})

	return out
}

// mergeTargets returns all issues except the current modal issue, for use as
// merge target candidates.
func (m model) mergeTargets() []github.Issue {
	var out []github.Issue
	for _, issue := range m.filteredIssues() {
		if issue.Number != m.modalIssue {
			out = append(out, issue)
		}
	}
	return out
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchIssuesCmd(m.settings.GetIssuesMax()), fetchTmuxStatusCmd, tmuxTickCmd(), refreshTickCmd())
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
