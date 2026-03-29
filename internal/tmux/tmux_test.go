package tmux

import (
	"testing"
)

func TestIsClaudeProcess(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"bash shell", "bash", false},
		{"zsh shell", "zsh", false},
		{"fish shell", "fish", false},
		{"sh shell", "sh", false},
		{"dash shell", "dash", false},
		{"ksh shell", "ksh", false},
		{"tcsh shell", "tcsh", false},
		{"csh shell", "csh", false},
		{"claude process", "claude", true},
		{"node process", "node", true},
		{"empty string", "", true},
		{"uppercase shell", "BASH", false},
		{"mixed case shell", "Zsh", false},
		{"shell with spaces", "  bash  ", false},
		{"python process", "python", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaudeProcess(tt.cmd)
			if got != tt.want {
				t.Errorf("isClaudeProcess(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestBuildClaudeCommand(t *testing.T) {
	tests := []struct {
		name            string
		prompt          string
		skipPermissions bool
		wantContains    string
		wantNotContains string
	}{
		{
			name:            "without skip permissions",
			prompt:          "Fix issue #42",
			skipPermissions: false,
			wantContains:    `claude "Fix issue #42"`,
			wantNotContains: "dangerously-skip-permissions",
		},
		{
			name:            "with skip permissions",
			prompt:          "Fix issue #42",
			skipPermissions: true,
			wantContains:    "--dangerously-skip-permissions",
		},
		{
			name:            "skip permissions includes IS_SANDBOX",
			prompt:          "test",
			skipPermissions: true,
			wantContains:    "IS_SANDBOX=1",
		},
		{
			name:         "prompt with special characters",
			prompt:       `Fix "bug" in code`,
			wantContains: `claude`,
		},
		{
			name:         "empty prompt",
			prompt:       "",
			wantContains: `claude ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildClaudeCommand(tt.prompt, tt.skipPermissions)
			if tt.wantContains != "" && !contains(got, tt.wantContains) {
				t.Errorf("BuildClaudeCommand(%q, %v) = %q, want it to contain %q",
					tt.prompt, tt.skipPermissions, got, tt.wantContains)
			}
			if tt.wantNotContains != "" && contains(got, tt.wantNotContains) {
				t.Errorf("BuildClaudeCommand(%q, %v) = %q, want it NOT to contain %q",
					tt.prompt, tt.skipPermissions, got, tt.wantNotContains)
			}
		})
	}
}

func TestIssueWindowRegex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMatch bool
		wantNum   string
	}{
		{"valid issue window", "issue-42-fix-bug", true, "42"},
		{"valid issue window with long name", "issue-123-some-long-title", true, "123"},
		{"issue-1-short", "issue-1-a", true, "1"},
		{"no match - no prefix", "fix-bug", false, ""},
		{"no match - wrong prefix", "pr-42-fix", false, ""},
		{"no match - no number", "issue-abc-fix", false, ""},
		{"no match - empty", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := issueWindowRe.FindStringSubmatch(tt.input)
			if tt.wantMatch {
				if m == nil {
					t.Errorf("expected %q to match issueWindowRe", tt.input)
					return
				}
				if m[1] != tt.wantNum {
					t.Errorf("expected issue number %q, got %q", tt.wantNum, m[1])
				}
			} else {
				if m != nil {
					t.Errorf("expected %q NOT to match issueWindowRe, but got %v", tt.input, m)
				}
			}
		})
	}
}

func TestPaneStatusConstants(t *testing.T) {
	if StatusWorking != 0 {
		t.Errorf("StatusWorking = %d, want 0", StatusWorking)
	}
	if StatusReview != 1 {
		t.Errorf("StatusReview = %d, want 1", StatusReview)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
