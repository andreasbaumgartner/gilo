package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var issueWindowRe = regexp.MustCompile(`^issue-(\d+)-`)

type Pane struct {
	WindowName string
	IssueNum   int
	LastLine   string
}

type WorktreeResult struct {
	Branch  string
	Path    string
	Session string
	Existed bool
	Reused  bool
	Split   bool
	Err     error
}

type ClaudeResult struct {
	Branch  string
	Path    string
	Existed bool
	Reused  bool
	Split   bool
	Err     error
}

func FetchStatus() []Pane {
	if _, err := exec.LookPath("tmux"); err != nil {
		return nil
	}

	var windowNames []string
	if os.Getenv("TMUX") != "" {
		out, err := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Output()
		if err == nil {
			windowNames = strings.Split(strings.TrimSpace(string(out)), "\n")
		}
	} else {
		out, err := exec.Command("tmux", "list-windows", "-a", "-F", "#{window_name}").Output()
		if err == nil {
			windowNames = strings.Split(strings.TrimSpace(string(out)), "\n")
		}
	}

	var panes []Pane
	for _, name := range windowNames {
		m := issueWindowRe.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		issueNum, _ := strconv.Atoi(m[1])

		var lastLine string
		out, err := exec.Command("tmux", "capture-pane", "-t", name, "-p").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				if l := strings.TrimSpace(lines[i]); l != "" {
					lastLine = l
					break
				}
			}
		}

		panes = append(panes, Pane{
			WindowName: name,
			IssueNum:   issueNum,
			LastLine:   lastLine,
		})
	}
	return panes
}

func WindowExists(branch string) bool {
	out, err := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Output()
	if err != nil {
		return false
	}
	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if name == branch {
			return true
		}
	}
	return false
}

func OpenWorktree(branch, worktreePath string, existed, split bool) WorktreeResult {
	if os.Getenv("TMUX") != "" {
		out, err := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Output()
		if err == nil {
			for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if name == branch {
					exec.Command("tmux", "select-window", "-t", branch).Run()
					return WorktreeResult{Branch: branch, Path: worktreePath, Existed: existed, Reused: true}
				}
			}
		}

		if split {
			tmuxErr := exec.Command("tmux", "split-window", "-h", "-d", "-c", worktreePath).Run()
			if tmuxErr != nil {
				return WorktreeResult{
					Err:    fmt.Errorf("worktree ready but tmux split-window failed: %w", tmuxErr),
					Branch: branch, Path: worktreePath, Existed: existed,
				}
			}
			return WorktreeResult{Branch: branch, Path: worktreePath, Existed: existed, Split: true}
		}

		tmuxErr := exec.Command("tmux", "new-window", "-d", "-n", branch, "-c", worktreePath).Run()
		if tmuxErr != nil {
			return WorktreeResult{
				Err:    fmt.Errorf("worktree ready but tmux new-window failed: %w", tmuxErr),
				Branch: branch, Path: worktreePath, Existed: existed,
			}
		}
		return WorktreeResult{Branch: branch, Path: worktreePath, Existed: existed}
	}

	if _, err := exec.LookPath("tmux"); err != nil {
		return WorktreeResult{
			Err:    fmt.Errorf("worktree ready at %s\nbut tmux is not installed", worktreePath),
			Branch: branch, Path: worktreePath, Existed: existed,
		}
	}

	sessionName := branch
	if exec.Command("tmux", "has-session", "-t", sessionName).Run() == nil {
		return WorktreeResult{Branch: branch, Path: worktreePath, Session: sessionName, Existed: existed, Reused: true}
	}

	if err := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", worktreePath).Run(); err != nil {
		return WorktreeResult{
			Err:    fmt.Errorf("worktree ready but failed to start tmux session: %w", err),
			Branch: branch, Path: worktreePath, Existed: existed,
		}
	}

	return WorktreeResult{Branch: branch, Path: worktreePath, Session: sessionName, Existed: existed}
}

func BuildClaudeCommand(prompt string, dangerouslySkipPermissions bool) string {
	cmd := "claude"
	if dangerouslySkipPermissions {
		cmd = "IS_SANDBOX=1 claude --dangerously-skip-permissions"
	}
	return fmt.Sprintf("%s %q", cmd, prompt)
}

func OpenClaude(branch, worktreePath string, existed bool, prompt string, split, skipPermissions bool) ClaudeResult {
	if os.Getenv("TMUX") == "" {
		if _, err := exec.LookPath("tmux"); err != nil {
			return ClaudeResult{
				Err:    fmt.Errorf("worktree ready at %s\nbut tmux is not installed", worktreePath),
				Branch: branch, Path: worktreePath, Existed: existed,
			}
		}
		if exec.Command("tmux", "has-session", "-t", branch).Run() == nil {
			return ClaudeResult{Branch: branch, Path: worktreePath, Existed: existed, Reused: true}
		}
		if err := exec.Command("tmux", "new-session", "-d", "-s", branch, "-c", worktreePath).Run(); err != nil {
			return ClaudeResult{
				Err:    fmt.Errorf("worktree ready but failed to start tmux session: %w", err),
				Branch: branch, Path: worktreePath, Existed: existed,
			}
		}
		cmd := BuildClaudeCommand(prompt, skipPermissions)
		exec.Command("tmux", "send-keys", "-t", branch, cmd, "Enter").Run()
		return ClaudeResult{Branch: branch, Path: worktreePath, Existed: existed}
	}

	// Inside tmux
	if WindowExists(branch) {
		return ClaudeResult{Branch: branch, Path: worktreePath, Existed: existed, Reused: true}
	}

	if split {
		if err := exec.Command("tmux", "split-window", "-h", "-d", "-c", worktreePath).Run(); err != nil {
			return ClaudeResult{
				Err:    fmt.Errorf("worktree ready but tmux split-window failed: %w", err),
				Branch: branch, Path: worktreePath, Existed: existed,
			}
		}
		cmd := BuildClaudeCommand(prompt, skipPermissions)
		exec.Command("tmux", "send-keys", "-t", ":.+", cmd, "Enter").Run()
		return ClaudeResult{Branch: branch, Path: worktreePath, Existed: existed, Split: true}
	}

	if err := exec.Command("tmux", "new-window", "-d", "-n", branch, "-c", worktreePath).Run(); err != nil {
		return ClaudeResult{
			Err:    fmt.Errorf("worktree ready but tmux new-window failed: %w", err),
			Branch: branch, Path: worktreePath, Existed: existed,
		}
	}
	cmd := BuildClaudeCommand(prompt, skipPermissions)
	exec.Command("tmux", "send-keys", "-t", branch, cmd, "Enter").Run()
	return ClaudeResult{Branch: branch, Path: worktreePath, Existed: existed}
}
