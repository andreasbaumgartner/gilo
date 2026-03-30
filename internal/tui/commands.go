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

func fetchLinkedPRsCmd(limit int) tea.Cmd {
	return func() tea.Msg {
		prs, err := github.FetchLinkedPRs(limit)
		return linkedPRsMsg{prs: prs, err: err}
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

func mergeIssuesCmd(source, target github.Issue) tea.Cmd {
	return func() tea.Msg {
		return issueMergedMsg{github.MergeIssues(source, target)}
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

func createClaudeTaskCmd(issue github.Issue, split bool, dangerouslySkipPermissions bool, additionalContext bool) tea.Cmd {
	return func() tea.Msg {
		branch, path, existed, err := worktree.Ensure(issue.Number, issue.Title)
		if err != nil {
			return claudeTaskCreatedMsg{tmux.ClaudeResult{Err: err, Branch: branch}}
		}
		prompt := buildClaudePrompt(issue, additionalContext)
		return claudeTaskCreatedMsg{tmux.OpenClaude(branch, path, existed, prompt, split, dangerouslySkipPermissions)}
	}
}

func buildClaudePrompt(issue github.Issue, additionalContext bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GitHub Issue #%d: %s", issue.Number, issue.Title))

	if additionalContext {
		sb.WriteString(fmt.Sprintf("\nAuthor: %s", issue.Author.Login))
		sb.WriteString(fmt.Sprintf("\nCreated: %s", issue.CreatedAt))
		sb.WriteString(fmt.Sprintf("\nState: %s", issue.State))
		if len(issue.Labels) > 0 {
			names := make([]string, len(issue.Labels))
			for i, l := range issue.Labels {
				names[i] = l.Name
			}
			sb.WriteString(fmt.Sprintf("\nLabels: %s", strings.Join(names, ", ")))
		}
	}

	if issue.Body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(issue.Body)
	}

	if additionalContext && len(issue.Comments) > 0 {
		sb.WriteString("\n\n--- Comments ---")
		for _, c := range issue.Comments {
			sb.WriteString(fmt.Sprintf("\n\n%s (%s):\n%s", c.Author.Login, c.CreatedAt, c.Body))
		}
	}

	sb.WriteString("\n\nPlease work on this issue.")
	return sb.String()
}
