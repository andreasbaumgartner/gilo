package selfupdate

import (
	"testing"
)

func TestInstallScriptURL(t *testing.T) {
	expected := "https://raw.githubusercontent.com/andreasbaumgartner/gilo/main/install.sh"
	if installScriptURL != expected {
		t.Errorf("installScriptURL = %q, want %q", installScriptURL, expected)
	}
}
