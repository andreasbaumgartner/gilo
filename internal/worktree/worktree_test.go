package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple lowercase", "hello world", "hello-world"},
		{"mixed case", "Hello World", "hello-world"},
		{"special characters", "fix: bug in API!", "fix-bug-in-api"},
		{"multiple spaces", "too   many   spaces", "too-many-spaces"},
		{"leading trailing special", "---hello---", "hello"},
		{"numbers preserved", "issue 42 fix", "issue-42-fix"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"long title truncated", "this-is-a-very-long-title-that-should-be-truncated-because-it-exceeds-the-limit", "this-is-a-very-long-title-that-should-be"},
		{"unicode characters", "café résumé", "caf-r-sum"},
		{"single word", "fix", "fix"},
		{"trailing dash after truncation", "abcdefghijklmnopqrstuvwxyzabcdefghijklmno-xyz", "abcdefghijklmnopqrstuvwxyzabcdefghijklmn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSlugifyMaxLength(t *testing.T) {
	long := "a"
	for i := 0; i < 100; i++ {
		long += "b"
	}
	got := Slugify(long)
	if len(got) > 40 {
		t.Errorf("Slugify produced string of length %d, want <= 40", len(got))
	}
}

func TestExists(t *testing.T) {
	t.Run("nonexistent path", func(t *testing.T) {
		if Exists("/nonexistent/path/that/does/not/exist") {
			t.Error("Exists returned true for nonexistent path")
		}
	})

	t.Run("regular directory without .git", func(t *testing.T) {
		dir := t.TempDir()
		if Exists(dir) {
			t.Error("Exists returned true for directory without .git file")
		}
	})

	t.Run("directory with .git directory (not worktree)", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".git"), 0755)
		if Exists(dir) {
			t.Error("Exists returned true for directory with .git directory (not a worktree)")
		}
	})

	t.Run("valid worktree with .git file", func(t *testing.T) {
		dir := t.TempDir()
		// Worktrees have a .git file (not directory) pointing to the main repo
		os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /some/path"), 0644)
		if !Exists(dir) {
			t.Error("Exists returned false for valid worktree directory")
		}
	})

	t.Run("file instead of directory", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "notadir")
		os.WriteFile(file, []byte("hello"), 0644)
		if Exists(file) {
			t.Error("Exists returned true for a file")
		}
	})
}
