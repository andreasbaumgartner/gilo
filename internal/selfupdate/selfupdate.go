package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const installScriptURL = "https://raw.githubusercontent.com/andreasbaumgartner/gilo/main/install.sh"

// Run downloads and executes the install script to update gilo, then restarts it.
func Run() error {
	fmt.Println("Updating gilo...")

	cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -fsSL %s | sh", installScriptURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Println("\nRestarting gilo...")

	binary, err := exec.LookPath("gilo")
	if err != nil {
		return fmt.Errorf("could not find gilo binary: %w", err)
	}

	return syscall.Exec(binary, os.Args[:1], os.Environ())
}
