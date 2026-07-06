// Package model defines the core data types for arsenal-ng.
//
// This file contains the data structures for cheat files (CheatFile, Action)
// and runtime types (Cheat) used throughout the application. It defines the
// structure of YAML cheat files and their runtime representations.
package model

// =============================================================================
// YAML File Structure
// =============================================================================

// CheatFile represents the structure of a YAML cheat file.
// Each file contains one tool with multiple actions.
type CheatFile struct {
	Tool    string   `yaml:"tool"`              // Tool name (e.g., "nmap", "ffuf")
	Tags    []string `yaml:"tags"`              // Tags (e.g., ["scan", "recon"])
	Stage   string   `yaml:"stage,omitempty"`   // Attack stage
	Tactic  string   `yaml:"tactic,omitempty"`  // MITRE ATT&CK tactic
	Actions []Action `yaml:"actions"`           // List of commands
}

// Action represents a single command entry in a cheat file.
type Action struct {
	Title   string   `yaml:"title"`              // Display title
	Desc    string   `yaml:"desc,omitempty"`     // Description
	Command string   `yaml:"command"`            // The actual command template
	Stage   string   `yaml:"stage,omitempty"`    // Per-action attack stage override
	Options []Option `yaml:"options,omitempty"`  // Pre-defined optional values for arguments
}

// =============================================================================
// Runtime Types
// =============================================================================

// ExecutionRecord represents a single execution history entry.
type ExecutionRecord struct {
	Tool      string   `json:"tool"`
	Title     string   `json:"title"`
	Command   string   `json:"command"`    // The final executed command with args filled
	Timestamp int64    `json:"timestamp"`  // Unix timestamp
	Args      []string `json:"args"`      // Argument values used
}

// Cheat is the runtime representation of a command.
// It's an enriched, flattened version of CheatFile + Action.
type Cheat struct {
	Tool     string   // Parent tool name
	Tags     []string // Tags associated with this cheat
	Stage    string   // Attack stage (inheritable from parent CheatFile or per-action)
	Tactic   string   // MITRE ATT&CK tactic or custom category
	Title    string   // Command title
	Desc     string   // Command description
	Command  string   // Command template with {{placeholders}}
	Options  []Option // Pre-defined argument options
	Filename string   // Source file path (for debugging)
	// Internal runtime field
	IsFavorite bool
}

// Argument represents a placeholder in a command template.
// Placeholders use the format {{name}} or {{name|default}}.

// Option represents a pre-defined value choice for an argument.
type Option struct {
	Arg   string `yaml:"arg"`   // Argument name this option applies to
	Value string `yaml:"value"` // Suggested value
}
type Argument struct {
	Name         string // Argument name (e.g., "ip", "port")
	DefaultValue string // Default value if specified with |
	Value        string // Current value (user input or default)
	Position     int    // Position in command string
}
