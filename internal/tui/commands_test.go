package tui

import (
	"strings"
	"testing"

	"github.com/andreasbaumgartner/gilo/internal/github"
)

func TestBuildClaudePrompt_Basic(t *testing.T) {
	issue := github.Issue{
		Number: 42,
		Title:  "Fix login bug",
		Body:   "The login page crashes on submit.",
	}
	issue.State = "OPEN"
	issue.Author.Login = "alice"
	issue.CreatedAt = "2026-01-15T10:00:00Z"

	prompt := buildClaudePrompt(issue, false)

	if !strings.Contains(prompt, "GitHub Issue #42: Fix login bug") {
		t.Error("prompt should contain issue number and title")
	}
	if !strings.Contains(prompt, "The login page crashes on submit.") {
		t.Error("prompt should contain issue body")
	}
	if !strings.Contains(prompt, "Please work on this issue.") {
		t.Error("prompt should end with action request")
	}

	// Should NOT contain additional context
	if strings.Contains(prompt, "Author:") {
		t.Error("basic prompt should not contain Author")
	}
	if strings.Contains(prompt, "Labels:") {
		t.Error("basic prompt should not contain Labels")
	}
	if strings.Contains(prompt, "Comments") {
		t.Error("basic prompt should not contain Comments")
	}
}

func TestBuildClaudePrompt_AdditionalContext(t *testing.T) {
	issue := github.Issue{
		Number: 42,
		Title:  "Fix login bug",
		Body:   "The login page crashes on submit.",
	}
	issue.State = "OPEN"
	issue.Author.Login = "alice"
	issue.CreatedAt = "2026-01-15T10:00:00Z"
	issue.Labels = []struct {
		Name string `json:"name"`
	}{
		{Name: "bug"},
		{Name: "urgent"},
	}
	issue.Comments = []struct {
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Body      string `json:"body"`
		CreatedAt string `json:"createdAt"`
	}{
		{
			Author:    struct{ Login string `json:"login"` }{Login: "bob"},
			Body:      "I can reproduce this.",
			CreatedAt: "2026-01-16T10:00:00Z",
		},
	}

	prompt := buildClaudePrompt(issue, true)

	if !strings.Contains(prompt, "GitHub Issue #42: Fix login bug") {
		t.Error("prompt should contain issue number and title")
	}
	if !strings.Contains(prompt, "Author: alice") {
		t.Error("prompt should contain author")
	}
	if !strings.Contains(prompt, "Created: 2026-01-15T10:00:00Z") {
		t.Error("prompt should contain created date")
	}
	if !strings.Contains(prompt, "State: OPEN") {
		t.Error("prompt should contain state")
	}
	if !strings.Contains(prompt, "Labels: bug, urgent") {
		t.Error("prompt should contain labels")
	}
	if !strings.Contains(prompt, "The login page crashes on submit.") {
		t.Error("prompt should contain issue body")
	}
	if !strings.Contains(prompt, "--- Comments ---") {
		t.Error("prompt should contain comments section")
	}
	if !strings.Contains(prompt, "bob (2026-01-16T10:00:00Z)") {
		t.Error("prompt should contain comment author and date")
	}
	if !strings.Contains(prompt, "I can reproduce this.") {
		t.Error("prompt should contain comment body")
	}
	if !strings.Contains(prompt, "Please work on this issue.") {
		t.Error("prompt should end with action request")
	}
}

func TestBuildClaudePrompt_AdditionalContextNoLabels(t *testing.T) {
	issue := github.Issue{
		Number: 10,
		Title:  "Simple task",
	}
	issue.State = "OPEN"
	issue.Author.Login = "carol"
	issue.CreatedAt = "2026-02-01T00:00:00Z"

	prompt := buildClaudePrompt(issue, true)

	if !strings.Contains(prompt, "Author: carol") {
		t.Error("prompt should contain author")
	}
	if strings.Contains(prompt, "Labels:") {
		t.Error("prompt should not contain Labels when there are none")
	}
	if strings.Contains(prompt, "--- Comments ---") {
		t.Error("prompt should not contain Comments section when there are none")
	}
}

func TestBuildClaudePrompt_EmptyBody(t *testing.T) {
	issue := github.Issue{
		Number: 5,
		Title:  "No body issue",
	}
	issue.State = "OPEN"
	issue.Author.Login = "dave"
	issue.CreatedAt = "2026-03-01T00:00:00Z"

	prompt := buildClaudePrompt(issue, false)

	if strings.Contains(prompt, "\n\n\n") {
		t.Error("prompt should not have extra blank lines when body is empty")
	}
	if !strings.Contains(prompt, "Please work on this issue.") {
		t.Error("prompt should still end with action request")
	}
}
