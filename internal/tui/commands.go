package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andreasbaumgartner/gilo/internal/github"
	"github.com/andreasbaumgartner/gilo/internal/tmux"
	"github.com/andreasbaumgartner/gilo/internal/worktree"
)

func tmuxTickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tmuxTickMsg{}
	})
}

func refreshTickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func fetchTmuxStatusCmd() tea.Msg {
	return tmuxStatusMsg(tmux.FetchStatus())
}

func fetchIssuesCmd(limit int) tea.Cmd {
	return func() tea.Msg {
		issues, err := github.FetchIssues(limit)
		if err != nil {
			return errMsg{err}
		}
		return issuesLoadedMsg(issues)
	}
}

func fetchLabelsCmd(issue github.Issue) tea.Cmd {
	return func() tea.Msg {
		labels, err := github.FetchLabels()
		if err != nil {
			return errMsg{err}
		}
		return labelsLoadedMsg{
			repoLabels:  labels,
			issueLabels: github.IssueLabelsMap(issue),
		}
	}
}

func applyLabelChangesCmd(issueNum int, toAdd, toRemove []string) tea.Cmd {
	return func() tea.Msg {
		err := github.ApplyLabelChanges(issueNum, toAdd, toRemove)
		return labelsUpdatedMsg{err: err}
	}
}

func closeIssueCmd(issueNum int) tea.Cmd {
	return func() tea.Msg {
		return issueClosedMsg{github.CloseIssue(issueNum)}
	}
}

func reopenIssueCmd(issueNum int) tea.Cmd {
	return func() tea.Msg {
		return issueReopenedMsg{github.ReopenIssue(issueNum)}
	}
}

func deleteIssueCmd(issueNum int) tea.Cmd {
	return func() tea.Msg {
		return issueDeletedMsg{github.DeleteIssue(issueNum)}
	}
}

func createWorktreeCmd(issue github.Issue, split bool) tea.Cmd {
	return func() tea.Msg {
		branch, path, existed, err := worktree.Ensure(issue.Number, issue.Title)
		if err != nil {
			return worktreeCreatedMsg{tmux.WorktreeResult{Err: err, Branch: branch}}
		}
		return worktreeCreatedMsg{tmux.OpenWorktree(branch, path, existed, split)}
	}
}

func createClaudeTaskCmd(issue github.Issue, split bool, dangerouslySkipPermissions bool) tea.Cmd {
	return func() tea.Msg {
		branch, path, existed, err := worktree.Ensure(issue.Number, issue.Title)
		if err != nil {
			return claudeTaskCreatedMsg{tmux.ClaudeResult{Err: err, Branch: branch}}
		}
		prompt := buildClaudePrompt(issue)
		return claudeTaskCreatedMsg{tmux.OpenClaude(branch, path, existed, prompt, split, dangerouslySkipPermissions)}
	}
}

func buildClaudePrompt(issue github.Issue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GitHub Issue #%d: %s", issue.Number, issue.Title))
	if issue.Body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(issue.Body)
	}
	sb.WriteString("\n\nPlease work on this issue.")
	return sb.String()
}
