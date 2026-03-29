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
