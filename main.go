package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	colorActive   = lipgloss.Color("62")
	colorInactive = lipgloss.Color("240")
	colorDim      = lipgloss.Color("240")
	colorYellow   = lipgloss.Color("3")
	colorGreen    = lipgloss.Color("2")
	colorRed      = lipgloss.Color("1")

	activeBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorActive)

	inactiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorInactive)

	titleStyle = lipgloss.NewStyle().
			Foreground(colorActive).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Background(colorActive).
			Foreground(lipgloss.Color("15")).
			Bold(true)

	dimStyle    = lipgloss.NewStyle().Foreground(colorDim)
	yellowStyle = lipgloss.NewStyle().Foreground(colorYellow)
	greenStyle  = lipgloss.NewStyle().Foreground(colorGreen)
	redStyle    = lipgloss.NewStyle().Foreground(colorRed)

	openBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("22")).
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Padding(0, 1)

	closedBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Padding(0, 1)

	keybindStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorActive).
			Padding(1, 2).
			Width(54)
)

type RepoLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Comments []struct {
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Body      string `json:"body"`
		CreatedAt string `json:"createdAt"`
	} `json:"comments"`
}

type focus int

const (
	focusList focus = iota
	focusDetail
)

type modalKind int

const (
	modalNone modalKind = iota
	modalBrowser
	modalComment
	modalWorktree
	modalCreate
	modalLabel
	modalClaudeTask
)

type issuesLoadedMsg []Issue
type errMsg struct{ err error }
type browserOpenedMsg struct{ err error }
type commentPostedMsg struct{ err error }
type issueCreatedMsg struct{ err error }
type labelsLoadedMsg struct {
	repoLabels  []RepoLabel
	issueLabels map[string]bool
}
type labelsUpdatedMsg struct{ err error }
type worktreeCreatedMsg struct {
	err     error
	branch  string
	path    string
	session string // non-empty if a new tmux session was created (not inside tmux)
	existed bool   // true if worktree already existed
	reused  bool   // true if an existing tmux session/window was reused
	split   bool   // true if opened as a split pane
}
type claudeTaskCreatedMsg struct {
	err     error
	branch  string
	path    string
	existed bool
	reused  bool
	split   bool // true if opened as a split pane
}

type TmuxPane struct {
	WindowName string
	IssueNum   int
	LastLine   string
}

type tmuxStatusMsg []TmuxPane
type tmuxTickMsg struct{}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	width  int
	height int

	focus  focus
	issues []Issue
	cursor int
	loaded bool
	err    error

	viewport viewport.Model

	modal       modalKind
	modalIssue  int
	modalStatus string
	textarea    textarea.Model

	textareaBody textarea.Model // body field for create-issue modal
	createFocus  int            // 0 = title, 1 = body

	repoLabels    []RepoLabel
	labelSelected map[string]bool
	labelOriginal map[string]bool
	labelCursor   int

	tmuxPanes []TmuxPane

	stateFilter string // "OPEN", "CLOSED", or "" (all)
}

func initialModel() model {
	w, h, _ := term.GetSize(0)
	vp := viewport.New(0, 0)
	return model{
		width:       w,
		height:      h,
		viewport:    vp,
		modal:       modalNone,
		stateFilter: "OPEN",
	}
}

func (m model) filteredIssues() []Issue {
	if m.stateFilter == "" {
		return m.issues
	}
	var out []Issue
	for _, issue := range m.issues {
		if issue.State == m.stateFilter {
			out = append(out, issue)
		}
	}
	return out
}

// ── Commands ──────────────────────────────────────────────────────────────────

var issueWindowRe = regexp.MustCompile(`^issue-(\d+)-`)

func tmuxTickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tmuxTickMsg{}
	})
}

func fetchTmuxStatus() tea.Msg {
	if _, err := exec.LookPath("tmux"); err != nil {
		return tmuxStatusMsg(nil)
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

	var panes []TmuxPane
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

		panes = append(panes, TmuxPane{
			WindowName: name,
			IssueNum:   issueNum,
			LastLine:   lastLine,
		})
	}
	return tmuxStatusMsg(panes)
}

func fetchIssues() tea.Msg {
	out, err := exec.Command("gh", "issue", "list",
		"--state", "all",
		"--json", "number,title,author,state,body,createdAt,labels,comments",
		"--limit", "50").Output()
	if err != nil {
		return errMsg{err}
	}
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return errMsg{err}
	}
	return issuesLoadedMsg(issues)
}

func ensureWorktree(issue Issue) (branch, worktreePath string, existed bool, err error) {
	branch = fmt.Sprintf("issue-%d-%s", issue.Number, slugify(issue.Title))

	rootOut, gitErr := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if gitErr != nil {
		return branch, "", false, fmt.Errorf("not a git repo: %w", gitErr)
	}
	repoRoot := strings.TrimSpace(string(rootOut))
	worktreePath = filepath.Join(filepath.Dir(repoRoot), ".worktrees", branch)

	existed = worktreeExists(worktreePath)
	if existed {
		return branch, worktreePath, true, nil
	}

	// Prune stale worktree references (e.g. manually deleted directories)
	exec.Command("git", "worktree", "prune").Run()

	// Remove leftover directory that isn't a valid worktree
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

func createWorktreeCmd(issue Issue, split bool) tea.Cmd {
	return func() tea.Msg {
		branch, worktreePath, existed, err := ensureWorktree(issue)
		if err != nil {
			return worktreeCreatedMsg{err: err, branch: branch}
		}
		return openWorktreeInTmux(branch, worktreePath, existed, split)
	}
}

func createClaudeTaskCmd(issue Issue, split bool) tea.Cmd {
	return func() tea.Msg {
		branch, worktreePath, existed, err := ensureWorktree(issue)
		if err != nil {
			return claudeTaskCreatedMsg{err: err, branch: branch}
		}
		return openClaudeInTmux(branch, worktreePath, existed, issue, split)
	}
}

func worktreeExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	// Verify it's a valid git worktree (has a .git file, not just a directory)
	gitPath := filepath.Join(path, ".git")
	gi, gerr := os.Stat(gitPath)
	return gerr == nil && !gi.IsDir()
}

func openWorktreeInTmux(branch, worktreePath string, existed bool, split bool) worktreeCreatedMsg {
	// Check if we're inside a tmux session
	if os.Getenv("TMUX") != "" {
		// Check if a window with this name already exists
		out, err := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Output()
		if err == nil {
			for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if name == branch {
					// Window exists — jump to it
					exec.Command("tmux", "select-window", "-t", branch).Run()
					return worktreeCreatedMsg{branch: branch, path: worktreePath, existed: existed, reused: true}
				}
			}
		}

		if split {
			// Create a split pane instead of a new window
			tmuxErr := exec.Command("tmux", "split-window", "-h", "-c", worktreePath).Run()
			if tmuxErr != nil {
				return worktreeCreatedMsg{
					err:    fmt.Errorf("worktree ready but tmux split-window failed: %w", tmuxErr),
					branch: branch, path: worktreePath, existed: existed,
				}
			}
			return worktreeCreatedMsg{branch: branch, path: worktreePath, existed: existed, split: true}
		}

		// No existing window — create a new one
		tmuxErr := exec.Command("tmux", "new-window", "-n", branch, "-c", worktreePath).Run()
		if tmuxErr != nil {
			return worktreeCreatedMsg{
				err:    fmt.Errorf("worktree ready but tmux new-window failed: %w", tmuxErr),
				branch: branch, path: worktreePath, existed: existed,
			}
		}
		return worktreeCreatedMsg{branch: branch, path: worktreePath, existed: existed}
	}

	// Not inside tmux — check if tmux is installed
	if _, err := exec.LookPath("tmux"); err != nil {
		return worktreeCreatedMsg{
			err:    fmt.Errorf("worktree ready at %s\nbut tmux is not installed", worktreePath),
			branch: branch, path: worktreePath, existed: existed,
		}
	}

	// Check if a tmux session with this name already exists
	sessionName := branch
	if exec.Command("tmux", "has-session", "-t", sessionName).Run() == nil {
		// Session exists — just report it for attaching
		return worktreeCreatedMsg{branch: branch, path: worktreePath, session: sessionName, existed: existed, reused: true}
	}

	// Create a new detached session
	if err := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", worktreePath).Run(); err != nil {
		return worktreeCreatedMsg{
			err:    fmt.Errorf("worktree ready but failed to start tmux session: %w", err),
			branch: branch, path: worktreePath, existed: existed,
		}
	}

	return worktreeCreatedMsg{branch: branch, path: worktreePath, session: sessionName, existed: existed}
}

func tmuxWindowExists(branch string) bool {
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

func openClaudeInTmux(branch, worktreePath string, existed bool, issue Issue, split bool) claudeTaskCreatedMsg {
	if os.Getenv("TMUX") == "" {
		if _, err := exec.LookPath("tmux"); err != nil {
			return claudeTaskCreatedMsg{
				err:    fmt.Errorf("worktree ready at %s\nbut tmux is not installed", worktreePath),
				branch: branch, path: worktreePath, existed: existed,
			}
		}
		// Outside tmux: check for existing session
		if exec.Command("tmux", "has-session", "-t", branch).Run() == nil {
			return claudeTaskCreatedMsg{branch: branch, path: worktreePath, existed: existed, reused: true}
		}
		// Create detached session, then send claude command
		if err := exec.Command("tmux", "new-session", "-d", "-s", branch, "-c", worktreePath).Run(); err != nil {
			return claudeTaskCreatedMsg{
				err:    fmt.Errorf("worktree ready but failed to start tmux session: %w", err),
				branch: branch, path: worktreePath, existed: existed,
			}
		}
		prompt := buildClaudePrompt(issue)
		exec.Command("tmux", "send-keys", "-t", branch, fmt.Sprintf("claude %q", prompt), "Enter").Run()
		return claudeTaskCreatedMsg{branch: branch, path: worktreePath, existed: existed}
	}

	// Inside tmux
	if tmuxWindowExists(branch) {
		return claudeTaskCreatedMsg{branch: branch, path: worktreePath, existed: existed, reused: true}
	}

	if split {
		// Create a split pane instead of a new window
		if err := exec.Command("tmux", "split-window", "-h", "-d", "-c", worktreePath).Run(); err != nil {
			return claudeTaskCreatedMsg{
				err:    fmt.Errorf("worktree ready but tmux split-window failed: %w", err),
				branch: branch, path: worktreePath, existed: existed,
			}
		}
		// Send claude command to the newly created pane (last pane)
		prompt := buildClaudePrompt(issue)
		exec.Command("tmux", "send-keys", "-t", ":.+", fmt.Sprintf("claude %q", prompt), "Enter").Run()
		return claudeTaskCreatedMsg{branch: branch, path: worktreePath, existed: existed, split: true}
	}

	// Create background window (-d = don't switch)
	if err := exec.Command("tmux", "new-window", "-d", "-n", branch, "-c", worktreePath).Run(); err != nil {
		return claudeTaskCreatedMsg{
			err:    fmt.Errorf("worktree ready but tmux new-window failed: %w", err),
			branch: branch, path: worktreePath, existed: existed,
		}
	}
	prompt := buildClaudePrompt(issue)
	exec.Command("tmux", "send-keys", "-t", branch, fmt.Sprintf("claude %q", prompt), "Enter").Run()
	return claudeTaskCreatedMsg{branch: branch, path: worktreePath, existed: existed}
}

func buildClaudePrompt(issue Issue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GitHub Issue #%d: %s", issue.Number, issue.Title))
	if issue.Body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(issue.Body)
	}
	sb.WriteString("\n\nPlease work on this issue.")
	return sb.String()
}

func slugify(s string) string {
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

func fetchLabelsCmd(issue Issue) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("gh", "label", "list",
			"--json", "name,color", "--limit", "100").Output()
		if err != nil {
			return errMsg{err}
		}
		var labels []RepoLabel
		if err := json.Unmarshal(out, &labels); err != nil {
			return errMsg{err}
		}
		issueLabels := make(map[string]bool)
		for _, l := range issue.Labels {
			issueLabels[l.Name] = true
		}
		return labelsLoadedMsg{repoLabels: labels, issueLabels: issueLabels}
	}
}

func applyLabelChanges(issueNum int, toAdd, toRemove []string) tea.Cmd {
	return func() tea.Msg {
		if len(toAdd) > 0 {
			if err := exec.Command("gh", "issue", "edit",
				fmt.Sprintf("%d", issueNum),
				"--add-label", strings.Join(toAdd, ",")).Run(); err != nil {
				return labelsUpdatedMsg{err: err}
			}
		}
		if len(toRemove) > 0 {
			if err := exec.Command("gh", "issue", "edit",
				fmt.Sprintf("%d", issueNum),
				"--remove-label", strings.Join(toRemove, ",")).Run(); err != nil {
				return labelsUpdatedMsg{err: err}
			}
		}
		return labelsUpdatedMsg{}
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchIssues, tmuxTickCmd())
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.detailInnerW()
		m.viewport.Height = m.detailInnerH()

	case tmuxTickMsg:
		return m, tea.Batch(fetchTmuxStatus, tmuxTickCmd())

	case tmuxStatusMsg:
		m.tmuxPanes = []TmuxPane(msg)
		if m.loaded && len(m.issues) > 0 {
			m.updateViewport()
		}

	case issuesLoadedMsg:
		m.issues = []Issue(msg)
		m.loaded = true
		filtered := m.filteredIssues()
		if m.cursor >= len(filtered) {
			m.cursor = max(0, len(filtered)-1)
		}
		m.updateViewport()

	case errMsg:
		m.err = msg.err

	case browserOpenedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
		}

	case commentPostedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssues
		}

	case issueCreatedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssues
		}

	case labelsLoadedMsg:
		m.repoLabels = msg.repoLabels
		m.labelSelected = make(map[string]bool)
		m.labelOriginal = make(map[string]bool)
		for k, v := range msg.issueLabels {
			m.labelSelected[k] = v
			m.labelOriginal[k] = v
		}
		m.modalStatus = ""

	case labelsUpdatedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.modal = modalNone
			m.modalStatus = ""
			return m, fetchIssues
		}

	case worktreeCreatedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			verb := "Created"
			if msg.existed {
				verb = "Opened existing"
			}
			if msg.session != "" {
				action := "Attach with: tmux attach -t " + msg.session
				if msg.reused {
					action = "Session already running — attach with:\ntmux attach -t " + msg.session
				}
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch:  %s\nPath:    %s\nSession: %s\n\n%s", verb, msg.branch, msg.path, msg.session, action)
			} else if msg.reused {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nJumped to existing tmux window.", verb, msg.branch, msg.path)
			} else if msg.split {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nOpened in tmux split pane.", verb, msg.branch, msg.path)
			} else {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nOpened in new tmux window.", verb, msg.branch, msg.path)
			}
		}

	case claudeTaskCreatedMsg:
		if msg.err != nil {
			m.modalStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			verb := "Created"
			if msg.existed {
				verb = "Opened existing"
			}
			if msg.reused {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nClaude already running in tmux window.", verb, msg.branch, msg.path)
			} else if msg.split {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nClaude started in tmux split pane.", verb, msg.branch, msg.path)
			} else {
				m.modalStatus = fmt.Sprintf("%s worktree!\n\nBranch: %s\nPath:   %s\n\nClaude started in background tmux window.", verb, msg.branch, msg.path)
			}
		}

	case tea.KeyMsg:
		if m.modal != modalNone {
			return m.updateModal(msg)
		}
		return m.updateNormal(msg)
	}

	return m, nil
}

func (m model) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case modalBrowser:
		if msg.String() == "esc" {
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalWorktree, modalClaudeTask:
		if msg.String() == "esc" {
			m.modal = modalNone
			m.modalStatus = ""
		}
		return m, nil

	case modalComment:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			m.modalStatus = ""
			return m, nil
		case "ctrl+d":
			body := strings.TrimSpace(m.textarea.Value())
			if body == "" {
				return m, nil
			}
			m.modalStatus = "Posting comment..."
			num := m.modalIssue
			return m, func() tea.Msg {
				err := exec.Command("gh", "issue", "comment",
					fmt.Sprintf("%d", num),
					"--body", body).Run()
				return commentPostedMsg{err}
			}
		}
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	case modalCreate:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			m.modalStatus = ""
			return m, nil
		case "tab":
			if m.createFocus == 0 {
				m.createFocus = 1
				m.textarea.Blur()
				m.textareaBody.Focus()
			} else {
				m.createFocus = 0
				m.textareaBody.Blur()
				m.textarea.Focus()
			}
			return m, nil
		case "ctrl+d":
			title := strings.TrimSpace(m.textarea.Value())
			if title == "" {
				return m, nil
			}
			body := strings.TrimSpace(m.textareaBody.Value())
			m.modalStatus = "Creating issue..."
			return m, func() tea.Msg {
				args := []string{"issue", "create", "--title", title}
				if body != "" {
					args = append(args, "--body", body)
				}
				err := exec.Command("gh", args...).Run()
				return issueCreatedMsg{err}
			}
		}
		var cmd tea.Cmd
		if m.createFocus == 0 {
			m.textarea, cmd = m.textarea.Update(msg)
		} else {
			m.textareaBody, cmd = m.textareaBody.Update(msg)
		}
		return m, cmd

	case modalLabel:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			m.modalStatus = ""
			return m, nil
		case "j", "down":
			if m.labelCursor < len(m.repoLabels)-1 {
				m.labelCursor++
			}
			return m, nil
		case "k", "up":
			if m.labelCursor > 0 {
				m.labelCursor--
			}
			return m, nil
		case " ", "enter":
			if len(m.repoLabels) > 0 {
				name := m.repoLabels[m.labelCursor].Name
				m.labelSelected[name] = !m.labelSelected[name]
			}
			return m, nil
		case "ctrl+d":
			var toAdd, toRemove []string
			for _, l := range m.repoLabels {
				was := m.labelOriginal[l.Name]
				now := m.labelSelected[l.Name]
				if now && !was {
					toAdd = append(toAdd, l.Name)
				} else if !now && was {
					toRemove = append(toRemove, l.Name)
				}
			}
			if len(toAdd) == 0 && len(toRemove) == 0 {
				m.modal = modalNone
				m.modalStatus = ""
				return m, nil
			}
			m.modalStatus = "Updating labels..."
			return m, applyLabelChanges(m.modalIssue, toAdd, toRemove)
		}
		return m, nil
	}
	return m, nil
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.focus == focusList {
			m.focus = focusDetail
		} else {
			m.focus = focusList
		}

	case "j", "down":
		if m.focus == focusList {
			filtered := m.filteredIssues()
			if m.cursor < len(filtered)-1 {
				m.cursor++
				m.updateViewport()
			}
		} else {
			m.viewport.ScrollDown(1)
		}

	case "k", "up":
		if m.focus == focusList {
			if m.cursor > 0 {
				m.cursor--
				m.updateViewport()
			}
		} else {
			m.viewport.ScrollUp(1)
		}

	case "enter", "o":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			m.modal = modalBrowser
			m.modalIssue = filtered[m.cursor].Number
			m.modalStatus = ""
			num := m.modalIssue
			return m, func() tea.Msg {
				err := exec.Command("gh", "issue", "view",
					fmt.Sprintf("%d", num), "--web").Run()
				return browserOpenedMsg{err}
			}
		}

	case "w", "W":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalWorktree
			m.modalIssue = issue.Number
			m.modalStatus = ""
			return m, createWorktreeCmd(issue, msg.String() == "W")
		}

	case "s", "S":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalClaudeTask
			m.modalIssue = issue.Number
			m.modalStatus = ""
			return m, createClaudeTaskCmd(issue, msg.String() == "S")
		}

	case "c":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			m.modal = modalComment
			m.modalIssue = filtered[m.cursor].Number
			m.modalStatus = ""
			ta := textarea.New()
			ta.Placeholder = "Write your comment..."
			ta.Focus()
			ta.SetWidth(48)
			ta.SetHeight(8)
			m.textarea = ta
		}

	case "n":
		m.modal = modalCreate
		m.modalStatus = ""
		m.createFocus = 0

		ta := textarea.New()
		ta.Placeholder = "Issue title..."
		ta.Focus()
		ta.SetWidth(48)
		ta.SetHeight(1)
		ta.CharLimit = 256
		m.textarea = ta

		tb := textarea.New()
		tb.Placeholder = "Describe the issue..."
		tb.Blur()
		tb.SetWidth(48)
		tb.SetHeight(6)
		m.textareaBody = tb

	case "l":
		filtered := m.filteredIssues()
		if len(filtered) > 0 {
			issue := filtered[m.cursor]
			m.modal = modalLabel
			m.modalIssue = issue.Number
			m.modalStatus = "Loading labels..."
			m.labelCursor = 0
			return m, fetchLabelsCmd(issue)
		}

	case "f":
		switch m.stateFilter {
		case "OPEN":
			m.stateFilter = "CLOSED"
		case "CLOSED":
			m.stateFilter = ""
		default:
			m.stateFilter = "OPEN"
		}
		filtered := m.filteredIssues()
		if m.cursor >= len(filtered) {
			m.cursor = max(0, len(filtered)-1)
		}
		m.updateViewport()
	}

	return m, nil
}

func (m *model) updateViewport() {
	m.viewport.Width = m.detailInnerW()
	m.viewport.Height = m.detailInnerH()
	m.viewport.SetContent(m.renderDetailContent())
	m.viewport.GotoTop()
}

// ── Layout helpers ────────────────────────────────────────────────────────────

func (m model) listW() int      { return m.width * 38 / 100 }
func (m model) detailW() int    { return m.width - m.listW() }
func (m model) mainH() int      { return m.height - 2 } // leave room for status bar
func (m model) listInnerW() int { return m.listW() - 4 }

func (m model) detailInnerW() int { return m.detailW() - 4 }
func (m model) detailInnerH() int { return m.mainH() - 2 }

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf(
			"\n  Error: %v\n\n  Make sure 'gh' is installed and authenticated.\n  Press q to quit.\n",
			m.err,
		)
	}

	left := m.renderList()
	right := m.renderDetail()
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	screen := body + "\n" + m.renderStatusBar()

	if m.modal != modalNone {
		modal := m.renderModal()
		screen = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)

		// Re-overlay status bar at bottom
		lines := strings.Split(screen, "\n")
		for len(lines) < m.height {
			lines = append(lines, "")
		}
		lines[m.height-1] = m.renderStatusBar()
		screen = strings.Join(lines, "\n")
	}

	return screen
}

// ── Modal ─────────────────────────────────────────────────────────────────────

func (m model) renderModal() string {
	switch m.modal {
	case modalBrowser:
		content := fmt.Sprintf("Opening issue #%d in browser...", m.modalIssue)
		if m.modalStatus != "" {
			content = m.modalStatus
		}
		hint := dimStyle.Render("esc dismiss")
		return modalStyle.Render(content + "\n\n" + hint)

	case modalWorktree:
		title := titleStyle.Render(fmt.Sprintf("Worktree for #%d", m.modalIssue))
		hint := dimStyle.Render("esc dismiss")
		content := "Setting up worktree..."
		if m.modalStatus != "" {
			content = m.modalStatus
		}
		return modalStyle.Render(strings.Join([]string{title, "", content, "", hint}, "\n"))

	case modalClaudeTask:
		title := titleStyle.Render(fmt.Sprintf("Claude Task for #%d", m.modalIssue))
		hint := dimStyle.Render("esc dismiss")
		content := "Setting up worktree and starting Claude..."
		if m.modalStatus != "" {
			content = m.modalStatus
		}
		return modalStyle.Render(strings.Join([]string{title, "", content, "", hint}, "\n"))

	case modalComment:
		title := titleStyle.Render(fmt.Sprintf("Comment on #%d", m.modalIssue))
		hint := dimStyle.Render("ctrl+d submit  │  esc cancel")
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else {
			body = m.textarea.View()
		}
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))

	case modalCreate:
		title := titleStyle.Render("Create New Issue")
		hint := dimStyle.Render("tab switch field  │  ctrl+d submit  │  esc cancel")
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else {
			body = strings.Join([]string{
				dimStyle.Render("Title:"),
				m.textarea.View(),
				"",
				dimStyle.Render("Body:"),
				m.textareaBody.View(),
			}, "\n")
		}
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))

	case modalLabel:
		title := titleStyle.Render(fmt.Sprintf("Labels for #%d", m.modalIssue))
		hint := dimStyle.Render("j/k navigate  │  space toggle  │  ctrl+d submit  │  esc cancel")
		var body string
		if m.modalStatus != "" {
			body = m.modalStatus
		} else if len(m.repoLabels) == 0 {
			body = dimStyle.Render("No labels found in this repository.")
		} else {
			var rows []string
			for i, l := range m.repoLabels {
				check := "[ ]"
				if m.labelSelected[l.Name] {
					check = "[x]"
				}
				labelColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#" + l.Color))
				line := fmt.Sprintf(" %s %s", check, labelColor.Render(l.Name))
				if i == m.labelCursor {
					line = selectedStyle.Render(fmt.Sprintf(" %s %s", check, l.Name))
				}
				rows = append(rows, line)
			}
			body = strings.Join(rows, "\n")
		}
		return modalStyle.Render(strings.Join([]string{title, "", body, "", hint}, "\n"))
	}
	return ""
}

// ── List panel ────────────────────────────────────────────────────────────────

func (m model) renderList() string {
	active := m.focus == focusList
	innerW := m.listInnerW()
	innerH := m.mainH() - 2

	filterLabel := "Open"
	if m.stateFilter == "CLOSED" {
		filterLabel = "Closed"
	} else if m.stateFilter == "" {
		filterLabel = "All"
	}

	header := titleStyle.Render("Issues") + " " + dimStyle.Render("["+filterLabel+"]")
	if !active {
		header = dimStyle.Render("Issues") + " " + dimStyle.Render("["+filterLabel+"]")
	}

	var rows []string
	rows = append(rows, header, "")

	filtered := m.filteredIssues()
	if !m.loaded {
		rows = append(rows, "  "+dimStyle.Render("Loading..."))
	} else if len(filtered) == 0 {
		rows = append(rows, "  "+dimStyle.Render("No "+strings.ToLower(filterLabel)+" issues found."))
	} else {
		for i, issue := range filtered {
			if len(rows) >= innerH {
				break
			}

			stateBadge := openBadge.Render("OPEN")
			if issue.State == "CLOSED" {
				stateBadge = closedBadge.Render("CLOSED")
			}

			tmuxIndicator := ""
			for _, p := range m.tmuxPanes {
				if p.IssueNum == issue.Number {
					tmuxIndicator = yellowStyle.Render(" ⟳")
					break
				}
			}

			// Fixed badge column: align titles regardless of OPEN/CLOSED
			pad := ""
			if issue.State != "CLOSED" {
				pad = "  " // extra spaces so OPEN aligns with CLOSED
			}

			num := dimStyle.Render(fmt.Sprintf("#%-4d", issue.Number))
			title := truncate(issue.Title, innerW-20)
			line := stateBadge + pad + " " + num + " " + title + tmuxIndicator

			if i == m.cursor {
				if active {
					rest := fmt.Sprintf("#%-4d %s", issue.Number, truncate(issue.Title, innerW-20))
					line = stateBadge + pad + " " + selectedStyle.Render(padRight(rest, innerW-12)) + tmuxIndicator
				} else {
					line = lipgloss.NewStyle().
						Foreground(lipgloss.Color("252")).
						Render(line)
				}
			}
			rows = append(rows, line)
		}
	}

	content := strings.Join(rows, "\n")
	return panelBorder(active).
		Width(m.listW()-2).
		Height(m.mainH()-2).
		Padding(0, 1).
		Render(content)
}

// ── Detail panel ─────────────────────────────────────────────────────────────

func (m model) renderDetail() string {
	active := m.focus == focusDetail
	return panelBorder(active).
		Width(m.detailW()-2).
		Height(m.mainH()-2).
		Padding(0, 1).
		Render(m.viewport.View())
}

func (m model) renderDetailContent() string {
	filtered := m.filteredIssues()
	if !m.loaded || len(filtered) == 0 {
		return dimStyle.Render("No issue selected.")
	}

	issue := filtered[m.cursor]
	w := m.detailInnerW()

	stateColor := greenStyle
	if issue.State == "CLOSED" {
		stateColor = redStyle
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render(fmt.Sprintf("#%d  %s", issue.Number, issue.Title)))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(w, 60)))
	b.WriteString("\n\n")

	// Meta
	b.WriteString(dimStyle.Render("State:   ") + stateColor.Render(issue.State) + "\n")
	b.WriteString(dimStyle.Render("Author:  ") + issue.Author.Login + "\n")
	b.WriteString(dimStyle.Render("Created: ") + issue.CreatedAt[:10] + "\n")

	if len(issue.Labels) > 0 {
		names := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			names[i] = l.Name
		}
		b.WriteString(dimStyle.Render("Labels:  ") + yellowStyle.Render(strings.Join(names, ", ")) + "\n")
	}

	// Tmux status
	for _, p := range m.tmuxPanes {
		if p.IssueNum == issue.Number {
			b.WriteString("\n")
			b.WriteString(dimStyle.Render("── Tmux ") + dimStyle.Render(strings.Repeat("─", max(0, w-10))) + "\n\n")
			b.WriteString("  " + yellowStyle.Render("Window: ") + p.WindowName + "\n")
			if p.LastLine != "" {
				b.WriteString("  " + yellowStyle.Render("Output: ") + truncate(p.LastLine, w-12) + "\n")
			}
			break
		}
	}

	// Body
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("── Description ") + dimStyle.Render(strings.Repeat("─", max(0, w-17))) + "\n\n")
	if issue.Body == "" {
		b.WriteString(dimStyle.Render("  (no description)\n"))
	} else {
		for _, line := range strings.Split(issue.Body, "\n") {
			b.WriteString("  " + line + "\n")
		}
	}

	// Comments
	if len(issue.Comments) > 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("── Comments (%d) ", len(issue.Comments))) +
			dimStyle.Render(strings.Repeat("─", max(0, w-20))) + "\n")

		for _, c := range issue.Comments {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(c.Author.Login))
			b.WriteString("  " + dimStyle.Render(c.CreatedAt[:10]) + "\n")
			for _, line := range strings.Split(c.Body, "\n") {
				b.WriteString("  " + line + "\n")
			}
		}
	}

	return b.String()
}

// ── Status bar ────────────────────────────────────────────────────────────────

func (m model) renderStatusBar() string {
	var keys []string
	switch m.modal {
	case modalComment, modalCreate:
		keys = []string{"tab switch field", "ctrl+d submit", "esc cancel"}
	case modalLabel:
		keys = []string{"j/k navigate", "space toggle", "ctrl+d submit", "esc cancel"}
	case modalBrowser, modalWorktree, modalClaudeTask:
		keys = []string{"esc dismiss"}
	default:
		keys = []string{
			"↑↓/jk navigate",
			"tab switch panel",
			"f filter state",
			"o open in browser",
			"c comment",
			"w/W worktree/split",
			"s/S claude/split",
			"l labels",
			"n new issue",
			"q quit",
		}
	}
	if m.modal == modalNone && len(m.tmuxPanes) > 0 {
		keys = append([]string{fmt.Sprintf("%d tmux ⟳", len(m.tmuxPanes))}, keys...)
	}
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = keybindStyle.Render(k)
	}
	bar := strings.Join(parts, " ")
	barW := lipgloss.Width(bar)
	if barW < m.width {
		bar += strings.Repeat(" ", m.width-barW)
	}
	return bar
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func panelBorder(active bool) lipgloss.Style {
	if active {
		return activeBorder
	}
	return inactiveBorder
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func padRight(s string, w int) string {
	n := w - lipgloss.Width(s)
	if n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Sprintf("Error: %v\n", err)
	}
}
