// Package ui provides the terminal user interface for arsenal-ng.
//
// This file defines all keyboard bindings used throughout the application
// including navigation keys, argument navigation keys, and control keys.
package ui

// =============================================================================
// Key Bindings - Centralized key definitions
// =============================================================================

const (
	// Navigation
	keyUp       = "up"
	keyDown     = "down"
	keyLeft     = "left"
	keyRight    = "right"
	keyCtrlP    = "ctrl+p"
	keyCtrlN    = "ctrl+n"
	keyPgUp     = "pgup"
	keyPgDown   = "pgdown"
	keyEnter    = "enter"
	keyEsc      = "esc"
	keyCtrlC    = "ctrl+c"
	keyHelp     = "?"
	keyQuit     = "q"

	// Vim-like navigation
	keyVimUp   = "k"
	keyVimDown = "j"
	keyVimSearch = "/"

	// Filter/favorites/history
	keyFilter     = "ctrl+f"
	keyFavorites  = "ctrl+e"
	keyHistory    = "ctrl+h"
	keyToggleFav  = "ctrl+d"

	// Argument navigation
	keyTab      = "tab"
	keyShiftTab = "shift+tab"
)

