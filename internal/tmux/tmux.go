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

// PaneStatus represents the state of a Claude session in a tmux pane.
type PaneStatus int

const (
	StatusWorking PaneStatus = iota
	StatusQuestion
	StatusReview
)

type Pane struct {
	WindowName string
	IssueNum   int
	LastLine   string
	Status     PaneStatus
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
		var capturedOutput string
		out, err := exec.Command("tmux", "capture-pane", "-t", name, "-p").Output()
		if err == nil {
			capturedOutput = string(out)
			lines := strings.Split(capturedOutput, "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				if l := strings.TrimSpace(lines[i]); l != "" {
					lastLine = l
					break
				}
			}
		}

		status := StatusWorking
		// Check if claude is still running in the pane by inspecting the
		// current command of the pane. If it is a shell (bash, zsh, etc.),
		// Claude has finished and the issue is ready for review.
		cmdOut, cmdErr := exec.Command("tmux", "list-panes", "-t", name, "-F", "#{pane_current_command}").Output()
		if cmdErr == nil {
			cmd := strings.TrimSpace(string(cmdOut))
			if !isClaudeProcess(cmd) {
				status = StatusReview
			} else if isClaudeIdle(capturedOutput) {
				// Claude Code is still running but showing the input
				// prompt — it is asking the user a question.
				status = StatusQuestion
			}
		}

		panes = append(panes, Pane{
			WindowName: name,
			IssueNum:   issueNum,
			LastLine:   lastLine,
			Status:     status,
		})
	}
	return panes
}

// isClaudeIdle checks if the captured pane content indicates that Claude Code
// has finished processing and is waiting for user input. When idle, Claude
// shows an input prompt box at the bottom of the terminal:
//
//	╭─────────────────────╮
//	│ >                   │
//	╰─────────────────────╯
func isClaudeIdle(output string) bool {
	if output == "" {
		return false
	}

	lines := strings.Split(output, "\n")

	// Collect last non-empty lines from the bottom of the pane.
	var lastLines []string
	for i := len(lines) - 1; i >= 0 && len(lastLines) < 5; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" {
			lastLines = append(lastLines, l)
		}
	}

	if len(lastLines) < 2 {
		return false
	}

	// The last non-empty line should be the bottom border of the input box.
	if !strings.HasPrefix(lastLines[0], "╰") || !strings.HasSuffix(lastLines[0], "╯") {
		return false
	}

	// One of the lines above should be a line inside the input box (bordered
	// by │ on both sides). We do NOT require ">" because Claude's permission
	// and question prompts use the same box but show "Yes / No" or other
	// content instead of a ">" caret.
	for _, line := range lastLines[1:] {
		if strings.HasPrefix(line, "│") && strings.HasSuffix(line, "│") {
			return true
		}
	}

	return false
}

// isClaudeProcess returns true if the command looks like an active Claude process
// rather than a shell prompt (which would indicate Claude has finished).
func isClaudeProcess(cmd string) bool {
	// When Claude is running, the pane command is typically "claude" or "node"
	// (since Claude Code is a Node.js app). When it finishes, the pane falls
	// back to the user's shell (bash, zsh, fish, etc.).
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	shells := []string{"bash", "zsh", "fish", "sh", "dash", "ksh", "tcsh", "csh"}
	for _, s := range shells {
		if cmd == s {
			return false
		}
	}
	return true
}

// SelectWindow switches tmux focus to the window with the given name.
// Returns true if the window was found and selected, false otherwise.
func SelectWindow(name string) bool {
	if os.Getenv("TMUX") == "" {
		return false
	}
	return exec.Command("tmux", "select-window", "-t", name).Run() == nil
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
