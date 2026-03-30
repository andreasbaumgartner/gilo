package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Settings struct {
	DangerouslySkipPermissions bool   `json:"dangerously_skip_permissions"`
	PermissionWarningAcked     bool   `json:"permission_warning_acked"`
	ColorScheme                string `json:"color_scheme,omitempty"`
	IssuesMax                  int    `json:"issues_max,omitempty"`
}

const DefaultIssuesMax = 25

var IssuesMaxOptions = []int{25, 50, 100, 200}

func (s Settings) GetIssuesMax() int {
	if s.IssuesMax <= 0 {
		return DefaultIssuesMax
	}
	return s.IssuesMax
}

func path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gilo", "settings.json")
}

func Load() Settings {
	data, err := os.ReadFile(path())
	if err != nil {
		return Settings{}
	}
	var s Settings
	json.Unmarshal(data, &s)
	return s
}

func Save(s Settings) error {
	p := path()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
