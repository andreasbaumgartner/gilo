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
