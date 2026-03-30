package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

type PR struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	State       string `json:"state"`       // "OPEN", "CLOSED", "MERGED"
	HeadRefName string `json:"headRefName"` // branch name
	Body        string `json:"body"`
}

// LinkedPRs maps issue numbers to PRs that reference them.
func FetchLinkedPRs(limit int) (map[int]PR, error) {
	out, err := exec.Command("gh", "pr", "list",
		"--state", "all",
		"--json", "number,title,state,headRefName,body",
		"--limit", fmt.Sprintf("%d", limit)).Output()
	if err != nil {
		return nil, err
	}
	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, err
	}

	return matchPRsToIssues(prs), nil
}

func matchPRsToIssues(prs []PR) map[int]PR {
	branchRe := regexp.MustCompile(`^issue-(\d+)-`)
	bodyRe := regexp.MustCompile(`(?i)(?:closes|close|resolves|resolve|fixes|fix|resolved)\s+#(\d+)`)
	linked := make(map[int]PR)

	for _, pr := range prs {
		// Match by branch name pattern (issue-{number}-*)
		if m := branchRe.FindStringSubmatch(pr.HeadRefName); m != nil {
			if num, err := strconv.Atoi(m[1]); err == nil {
				if existing, exists := linked[num]; !exists || pr.State == "OPEN" && existing.State != "OPEN" {
					linked[num] = pr
				}
			}
		}
		// Match by body references (closes #N, resolves #N, fixes #N)
		for _, m := range bodyRe.FindAllStringSubmatch(pr.Body, -1) {
			if num, err := strconv.Atoi(m[1]); err == nil {
				if existing, exists := linked[num]; !exists || pr.State == "OPEN" && existing.State != "OPEN" {
					linked[num] = pr
				}
			}
		}
	}
	return linked
}

func FetchIssues(limit int) ([]Issue, error) {
	out, err := exec.Command("gh", "issue", "list",
		"--state", "all",
		"--json", "number,title,author,state,body,createdAt,labels,comments",
		"--limit", fmt.Sprintf("%d", limit)).Output()
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func FetchLabels() ([]RepoLabel, error) {
	out, err := exec.Command("gh", "label", "list",
		"--json", "name,color", "--limit", "100").Output()
	if err != nil {
		return nil, err
	}
	var labels []RepoLabel
	if err := json.Unmarshal(out, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

func IssueLabelsMap(issue Issue) map[string]bool {
	m := make(map[string]bool)
	for _, l := range issue.Labels {
		m[l.Name] = true
	}
	return m
}

func ApplyLabelChanges(issueNum int, toAdd, toRemove []string) error {
	if len(toAdd) > 0 {
		if err := exec.Command("gh", "issue", "edit",
			fmt.Sprintf("%d", issueNum),
			"--add-label", strings.Join(toAdd, ",")).Run(); err != nil {
			return err
		}
	}
	if len(toRemove) > 0 {
		if err := exec.Command("gh", "issue", "edit",
			fmt.Sprintf("%d", issueNum),
			"--remove-label", strings.Join(toRemove, ",")).Run(); err != nil {
			return err
		}
	}
	return nil
}

func PostComment(issueNum int, body string) error {
	return exec.Command("gh", "issue", "comment",
		fmt.Sprintf("%d", issueNum),
		"--body", body).Run()
}

func CreateIssue(title, body string) error {
	args := []string{"issue", "create", "--title", title}
	if body != "" {
		args = append(args, "--body", body)
	}
	return exec.Command("gh", args...).Run()
}

func CloseIssue(issueNum int) error {
	return exec.Command("gh", "issue", "close",
		fmt.Sprintf("%d", issueNum)).Run()
}

func ReopenIssue(issueNum int) error {
	return exec.Command("gh", "issue", "reopen",
		fmt.Sprintf("%d", issueNum)).Run()
}

func DeleteIssue(issueNum int) error {
	return exec.Command("gh", "issue", "delete",
		fmt.Sprintf("%d", issueNum), "--yes").Run()
}

func OpenInBrowser(issueNum int) error {
	return exec.Command("gh", "issue", "view",
		fmt.Sprintf("%d", issueNum), "--web").Run()
}
