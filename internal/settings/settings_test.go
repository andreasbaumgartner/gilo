package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultSettings(t *testing.T) {
	// Override HOME to a temp directory so we don't read real settings
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	s := Load()
	if s.DangerouslySkipPermissions {
		t.Error("default DangerouslySkipPermissions should be false")
	}
	if s.PermissionWarningAcked {
		t.Error("default PermissionWarningAcked should be false")
	}
}

func TestSaveAndLoad(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	want := Settings{
		DangerouslySkipPermissions: true,
		PermissionWarningAcked:     true,
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got := Load()
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	err := Save(Settings{})
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	settingsDir := filepath.Join(tmpHome, ".gilo")
	if _, err := os.Stat(settingsDir); os.IsNotExist(err) {
		t.Error("Save() did not create .gilo directory")
	}
}

func TestSaveWritesValidJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	s := Settings{DangerouslySkipPermissions: true}
	if err := Save(s); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpHome, ".gilo", "settings.json"))
	if err != nil {
		t.Fatalf("could not read settings file: %v", err)
	}

	var loaded Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("settings file is not valid JSON: %v", err)
	}
	if loaded.DangerouslySkipPermissions != true {
		t.Error("saved JSON does not contain correct DangerouslySkipPermissions value")
	}
}

func TestAdditionalContextSaveAndLoad(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	want := Settings{
		AdditionalContext: true,
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got := Load()
	if got.AdditionalContext != true {
		t.Error("expected AdditionalContext to be true after save and load")
	}
}

func TestPlanModeSaveAndLoad(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	want := Settings{
		PlanMode: true,
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got := Load()
	if got.PlanMode != true {
		t.Error("expected PlanMode to be true after save and load")
	}
}

func TestLoadWithCorruptFile(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	settingsDir := filepath.Join(tmpHome, ".gilo")
	os.MkdirAll(settingsDir, 0755)
	os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("not json"), 0644)

	s := Load()
	// Should return zero-value settings on parse error
	if s.DangerouslySkipPermissions || s.PermissionWarningAcked {
		t.Error("Load() with corrupt file should return default settings")
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	s1 := Settings{DangerouslySkipPermissions: true}
	Save(s1)

	s2 := Settings{DangerouslySkipPermissions: false, PermissionWarningAcked: true}
	Save(s2)

	got := Load()
	if got != s2 {
		t.Errorf("after overwrite, Load() = %+v, want %+v", got, s2)
	}
}
