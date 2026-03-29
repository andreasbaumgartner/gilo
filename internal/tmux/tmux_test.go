package tmux

import "testing"

func TestIsClaudeIdle(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "empty output",
			output: "",
			want:   false,
		},
		{
			name:   "claude actively working",
			output: "⏺ Editing file internal/tmux/tmux.go\n  Writing content...\n",
			want:   false,
		},
		{
			name: "claude idle with input prompt",
			output: "some previous output\n" +
				"╭──────────────────────────────────────────────────────╮\n" +
				"│ >                                                    │\n" +
				"╰──────────────────────────────────────────────────────╯\n" +
				"\n",
			want: true,
		},
		{
			name: "claude idle with text in prompt",
			output: "some output\n" +
				"╭──────────────────────────────────────────────────────╮\n" +
				"│ > do something                                       │\n" +
				"╰──────────────────────────────────────────────────────╯\n",
			want: true,
		},
		{
			name:   "shell prompt not matching",
			output: "user@host:~$ \n",
			want:   false,
		},
		{
			name: "box characters in output but not input prompt",
			output: "╭── Error ──────────╮\n" +
				"│ Something failed  │\n" +
				"╰──────────────────╯\n" +
				"⏺ Retrying...\n",
			want: false,
		},
		{
			name: "only bottom border no prompt line",
			output: "some text\n" +
				"╰──────────────────────────────────────────────────────╯\n",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaudeIdle(tt.output)
			if got != tt.want {
				t.Errorf("isClaudeIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsClaudeProcess(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"bash", false},
		{"zsh", false},
		{"fish", false},
		{"claude", true},
		{"node", true},
		{"sh", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := isClaudeProcess(tt.cmd)
			if got != tt.want {
				t.Errorf("isClaudeProcess(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}
