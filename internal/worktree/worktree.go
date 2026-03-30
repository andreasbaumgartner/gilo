package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// WorktreeEntry represents a single git worktree.
type WorktreeEntry struct {
	Path   string
	Branch string
	Head   string
	Bare   bool
}

// List returns all git worktrees by parsing `git worktree list --porcelain`.
func List() ([]WorktreeEntry, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var entries []WorktreeEntry
	var current WorktreeEntry
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = WorktreeEntry{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.Bare = true
		case line == "":
			if current.Path != "" {
				entries = append(entries, current)
				current = WorktreeEntry{}
			}
		}
	}
	return entries, nil
}

// Remove removes a git worktree by path, using force if necessary.
func Remove(worktreePath string) error {
	_, err := exec.Command("git", "worktree", "remove", worktreePath).CombinedOutput()
	if err != nil {
		// Try force removal if normal removal fails
		out, err2 := exec.Command("git", "worktree", "remove", "--force", worktreePath).CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("git worktree remove failed: %s", strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func Slugify(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
		s = strings.TrimRight(s, "-")
	}
	return s
}

func Exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	gitPath := filepath.Join(path, ".git")
	gi, gerr := os.Stat(gitPath)
	return gerr == nil && !gi.IsDir()
}

func checkWritable(dir string) error {
	tmp, err := os.CreateTemp(dir, ".gilo-check-*")
	if err != nil {
		return err
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return nil
}

func Ensure(issueNumber int, title string) (branch, worktreePath string, existed bool, err error) {
	branch = fmt.Sprintf("issue-%d-%s", issueNumber, Slugify(title))

	rootOut, gitErr := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if gitErr != nil {
		return branch, "", false, fmt.Errorf("not a git repo: %w", gitErr)
	}
	repoRoot := strings.TrimSpace(string(rootOut))

	refsDir := filepath.Join(repoRoot, ".git", "refs", "heads")
	if err := checkWritable(refsDir); err != nil {
		gitDir := filepath.Join(repoRoot, ".git")
		return branch, "", false, fmt.Errorf(
			"permission denied: cannot write to %s\nFix with: sudo chown -R $(id -u):$(id -g) %s",
			refsDir, gitDir)
	}

	worktreePath = filepath.Join(filepath.Dir(repoRoot), ".worktrees", branch)

	existed = Exists(worktreePath)
	if existed {
		return branch, worktreePath, true, nil
	}

	exec.Command("git", "worktree", "prune").Run()

	if info, serr := os.Stat(worktreePath); serr == nil && info.IsDir() {
		os.RemoveAll(worktreePath)
	}

	os.MkdirAll(filepath.Dir(worktreePath), 0o755)

	out, addErr := exec.Command("git", "worktree", "add", "-b", branch, worktreePath).CombinedOutput()
	if addErr != nil {
		out2, addErr2 := exec.Command("git", "worktree", "add", worktreePath, branch).CombinedOutput()
		if addErr2 != nil {
			return branch, "", false, fmt.Errorf("%s\n%s", string(out), string(out2))
		}
	}
	return branch, worktreePath, false, nil
}
