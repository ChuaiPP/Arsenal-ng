//go:build darwin || linux

// Package output provides terminal output functionality.
//
// This file uses TIOCSTI ioctl to inject commands into the terminal's input
// buffer, making them appear as if the user typed them. This allows users to
// review and edit commands before execution. Supports both Linux and macOS.
package output

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// =============================================================================
// Platform-specific ioctl Constants
// =============================================================================

const (
	// TIOCSTI - Terminal I/O Control Simulate Terminal Input
	// Injects characters into terminal input buffer
	ioctlTIOCSTI_Linux  = 0x5412
	ioctlTIOCSTI_Darwin = 0x80017472

	// TCGETS/TCSETS - Get/Set terminal attributes
	ioctlTCGETS_Linux  = 0x5401
	ioctlTCSETS_Linux  = 0x5402
	ioctlTCGETS_Darwin = 0x40487413
	ioctlTCSETS_Darwin = 0x80487414
)

// =============================================================================
// Terminal Prefill
// =============================================================================

// ToTerminal writes a command to the terminal's input buffer.
// The command appears as if the user typed it, ready for editing.
// This allows users to review and modify the command before execution.
//
// Note: On Linux kernel 6.2+, this requires:
//
//	sysctl -w dev.tty.legacy_tiocsti=1

func ToTerminal(command string) {
	if len(command) == 0 {
		log.Printf("WARNING: Attempted to output empty command to terminal")
		return
	}

	// Strategy 1: Try TIOCSTI first (traditional approach)
	if tryTIOCSTI(command) {
		log.Printf("Command injected via TIOCSTI (%d bytes)", len(command))
		return
	}

	log.Printf("TIOCSTI failed, trying clipboard fallback...")

	// Strategy 2: Try clipboard (xclip / wl-copy / pbcopy)
	if tryClipboard(command) {
		fmt.Printf("\nCommand copied to clipboard. Paste with Ctrl+Shift+V or right-click.\n")
		log.Printf("Command copied to clipboard (%d bytes)", len(command))
		return
	}

	// Strategy 3: Last resort - print to stdout
	fmt.Printf("\n%s\n", separatorLine())
	fmt.Printf("Copy this command:\n%s\n", command)
	fmt.Printf("%s\n", separatorLine())
}

// tryTIOCSTI attempts to inject the command via TIOCSTI ioctl.
// Returns true on success, false if any step fails.
func tryTIOCSTI(command string) bool {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		log.Printf("TIOCSTI: cannot open /dev/tty: %v", err)
		return false
	}
	defer tty.Close()

	fd := int(tty.Fd())
	tiocsti, tcgets, tcsets := getPlatformIoctls()

	oldTermios, err := unix.IoctlGetTermios(fd, tcgets)
	if err != nil {
		log.Printf("TIOCSTI: cannot get termios: %v", err)
		return false
	}

	newTermios := *oldTermios
	newTermios.Lflag &^= unix.ECHO
	newTermios.Lflag &^= unix.ICANON
	if err := unix.IoctlSetTermios(fd, tcsets, &newTermios); err != nil {
		log.Printf("TIOCSTI: cannot set termios: %v", err)
		return false
	}

	for _, char := range []byte(command) {
		_, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			uintptr(fd),
			uintptr(tiocsti),
			uintptr(unsafe.Pointer(&char)),
		)
		if errno != 0 {
			log.Printf("TIOCSTI: ioctl failed with errno %d", errno)
			_ = unix.IoctlSetTermios(fd, tcsets, oldTermios)
			return false
		}
	}

	_ = unix.IoctlSetTermios(fd, tcsets, oldTermios)
	return true
}

// tryClipboard copies the command to the system clipboard.
func tryClipboard(command string) bool {
	clipCmd := findClipboardCommand()
	if clipCmd == "" {
		return false
	}

	cmd := exec.Command("sh", "-c", clipCmd)
	cmd.Stdin = strings.NewReader(command)
	if err := cmd.Run(); err != nil {
		log.Printf("Clipboard: failed to run %s: %v", clipCmd, err)
		return false
	}
	return true
}

// findClipboardCommand returns a shell pipeline to copy stdin to clipboard.
func findClipboardCommand() string {
	// Wayland
	if _, err := exec.LookPath("wl-copy"); err == nil {
		return "wl-copy"
	}
	// X11 primary clipboard tools
	if _, err := exec.LookPath("xclip"); err == nil {
		return "xclip -selection clipboard"
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		return "xsel --clipboard --input"
	}
	// macOS
	if _, err := exec.LookPath("pbcopy"); err == nil {
		return "pbcopy"
	}
	return ""
}

// separatorLine returns a visible separator for stdout fallback.
func separatorLine() string {
	return strings.Repeat("=", 72)
}

// getPlatformIoctls returns the correct ioctl constants for the current OS.
func getPlatformIoctls() (tiocsti, tcgets, tcsets uint) {
	if runtime.GOOS == "darwin" {
		return ioctlTIOCSTI_Darwin, ioctlTCGETS_Darwin, ioctlTCSETS_Darwin
	}
	return ioctlTIOCSTI_Linux, ioctlTCGETS_Linux, ioctlTCSETS_Linux
}
