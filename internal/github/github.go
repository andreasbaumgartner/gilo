package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
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

func FetchIssues() ([]Issue, error) {
	out, err := exec.Command("gh", "issue", "list",
		"--state", "all",
		"--json", "number,title,author,state,body,createdAt,labels,comments",
		"--limit", "50").Output()
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

func DeleteIssue(issueNum int) error {
	return exec.Command("gh", "issue", "delete",
		fmt.Sprintf("%d", issueNum), "--yes").Run()
}

func OpenInBrowser(issueNum int) error {
	return exec.Command("gh", "issue", "view",
		fmt.Sprintf("%d", issueNum), "--web").Run()
}
