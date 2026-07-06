//go:build windows

// Package output provides terminal output functionality.
//
// This file implements the Windows version using clipboard copy.
// On Windows, TIOCSTI is not available, so we copy the command
// to the clipboard and inform the user.
package output

import (
	"fmt"
	"log"
	"os/exec"
)

// ToTerminal copies the command to the clipboard on Windows.
// The command is placed in the clipboard for pasting.
func ToTerminal(command string) {
	if len(command) == 0 {
		log.Printf("WARNING: Attempted to output empty command to terminal")
		return
	}

	// Use PowerShell to set clipboard (most reliable on Windows)
	psCmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Set-Clipboard", "-Value", command)
	if err := psCmd.Run(); err != nil {
		log.Printf("ERROR: Failed to copy to clipboard: %v", err)
		fmt.Printf("\nCommand (clipboard copy failed, here it is):\n%s\n", command)
		return
	}

	fmt.Printf("\nCommand copied to clipboard! Paste with Ctrl+V\n")
	log.Printf("Command copied to clipboard (%d bytes)", len(command))
}
