package github

import (
	"testing"
)

func TestIssueLabelsMap(t *testing.T) {
	tests := []struct {
		name   string
		issue  Issue
		want   map[string]bool
		length int
	}{
		{
			name:   "empty labels",
			issue:  Issue{},
			want:   map[string]bool{},
			length: 0,
		},
		{
			name: "single label",
			issue: Issue{
				Labels: []struct {
					Name string `json:"name"`
				}{
					{Name: "bug"},
				},
			},
			want:   map[string]bool{"bug": true},
			length: 1,
		},
		{
			name: "multiple labels",
			issue: Issue{
				Labels: []struct {
					Name string `json:"name"`
				}{
					{Name: "bug"},
					{Name: "enhancement"},
					{Name: "help wanted"},
				},
			},
			want:   map[string]bool{"bug": true, "enhancement": true, "help wanted": true},
			length: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IssueLabelsMap(tt.issue)
			if len(got) != tt.length {
				t.Errorf("IssueLabelsMap() returned map with %d entries, want %d", len(got), tt.length)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("IssueLabelsMap()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestMatchPRsToIssues(t *testing.T) {
	tests := []struct {
		name string
		prs  []PR
		want map[int]PR
	}{
		{
			name: "match by branch name",
			prs: []PR{
				{Number: 10, Title: "Fix login", State: "OPEN", HeadRefName: "issue-5-fix-login"},
			},
			want: map[int]PR{
				5: {Number: 10, Title: "Fix login", State: "OPEN", HeadRefName: "issue-5-fix-login"},
			},
		},
		{
			name: "match by body closes reference",
			prs: []PR{
				{Number: 20, Title: "Add feature", State: "OPEN", HeadRefName: "feature-branch", Body: "Closes #7"},
			},
			want: map[int]PR{
				7: {Number: 20, Title: "Add feature", State: "OPEN", HeadRefName: "feature-branch", Body: "Closes #7"},
			},
		},
		{
			name: "match by body resolves reference",
			prs: []PR{
				{Number: 30, Title: "Fix bug", State: "MERGED", HeadRefName: "bugfix", Body: "resolves #12"},
			},
			want: map[int]PR{
				12: {Number: 30, Title: "Fix bug", State: "MERGED", HeadRefName: "bugfix", Body: "resolves #12"},
			},
		},
		{
			name: "match by body fixes reference",
			prs: []PR{
				{Number: 40, Title: "Hotfix", State: "OPEN", HeadRefName: "hotfix", Body: "fixes #3"},
			},
			want: map[int]PR{
				3: {Number: 40, Title: "Hotfix", State: "OPEN", HeadRefName: "hotfix", Body: "fixes #3"},
			},
		},
		{
			name: "open PR takes priority over merged",
			prs: []PR{
				{Number: 50, Title: "Old fix", State: "MERGED", HeadRefName: "issue-8-old-fix"},
				{Number: 60, Title: "New fix", State: "OPEN", HeadRefName: "issue-8-new-fix"},
			},
			want: map[int]PR{
				8: {Number: 60, Title: "New fix", State: "OPEN", HeadRefName: "issue-8-new-fix"},
			},
		},
		{
			name: "no match for unrelated branch",
			prs: []PR{
				{Number: 70, Title: "Random", State: "OPEN", HeadRefName: "feature-xyz"},
			},
			want: map[int]PR{},
		},
		{
			name: "multiple issues referenced in body",
			prs: []PR{
				{Number: 80, Title: "Big fix", State: "OPEN", HeadRefName: "multi-fix", Body: "Closes #1\nFixes #2"},
			},
			want: map[int]PR{
				1: {Number: 80, Title: "Big fix", State: "OPEN", HeadRefName: "multi-fix", Body: "Closes #1\nFixes #2"},
				2: {Number: 80, Title: "Big fix", State: "OPEN", HeadRefName: "multi-fix", Body: "Closes #1\nFixes #2"},
			},
		},
		{
			name: "resolved keyword matches",
			prs: []PR{
				{Number: 90, Title: "Fix it", State: "OPEN", HeadRefName: "fix", Body: "-resolved #42"},
			},
			want: map[int]PR{
				42: {Number: 90, Title: "Fix it", State: "OPEN", HeadRefName: "fix", Body: "-resolved #42"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPRsToIssues(tt.prs)
			if len(got) != len(tt.want) {
				t.Errorf("matchPRsToIssues() returned %d entries, want %d", len(got), len(tt.want))
			}
			for issueNum, wantPR := range tt.want {
				gotPR, ok := got[issueNum]
				if !ok {
					t.Errorf("expected issue #%d to have a linked PR", issueNum)
					continue
				}
				if gotPR.Number != wantPR.Number {
					t.Errorf("issue #%d: got PR #%d, want PR #%d", issueNum, gotPR.Number, wantPR.Number)
				}
				if gotPR.State != wantPR.State {
					t.Errorf("issue #%d: got state %q, want %q", issueNum, gotPR.State, wantPR.State)
				}
			}
		})
	}
}

func TestIssueLabelsMapLookup(t *testing.T) {
	issue := Issue{
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "bug"},
		},
	}
	m := IssueLabelsMap(issue)

	if !m["bug"] {
		t.Error("expected 'bug' to be in the map")
	}
	if m["nonexistent"] {
		t.Error("expected 'nonexistent' to be false (zero value)")
	}
}
