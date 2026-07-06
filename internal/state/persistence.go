// Package state manages persistent and session-scoped state for the application.
//
// This file provides persistence for execution history and favorites
// using JSON files in the arsenal-ng config directory.
package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/halilkirazkaya/arsenal-ng/internal/config"
	"github.com/halilkirazkaya/arsenal-ng/internal/model"
)

// =============================================================================
// Persistence Manager
// =============================================================================

// Persistence manages history and favorites on disk.
type Persistence struct {
	configDir string
	maxHistory int
}

// NewPersistence creates a new persistence manager.
func NewPersistence() (*Persistence, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config dir: %w", err)
	}
	return &Persistence{configDir: dir, maxHistory: 500}, nil
}

// =============================================================================
// Execution History
// =============================================================================

// AddHistory appends an execution record to the history file.
func (p *Persistence) AddHistory(record model.ExecutionRecord) error {
	records, _ := p.LoadHistory()
	records = append(records, record)
	// Trim to max size
	if len(records) > p.maxHistory {
		records = records[len(records)-p.maxHistory:]
	}
	return p.saveJSON("history.json", records)
}

// LoadHistory loads all execution history records.
func (p *Persistence) LoadHistory() ([]model.ExecutionRecord, error) {
	var records []model.ExecutionRecord
	err := p.loadJSON("history.json", &records)
	return records, err
}

// =============================================================================
// Favorites
// =============================================================================

// FavoriteEntry stores a favorited cheat reference.
type FavoriteEntry struct {
	Tool    string `json:"tool"`
	Title   string `json:"title"`
	Command string `json:"command"`
}

// AddFavorite adds a cheat to favorites.
func (p *Persistence) AddFavorite(cheat *model.Cheat) error {
	favs, _ := p.LoadFavorites()
	// Check for duplicates
	for _, f := range favs {
		if f.Tool == cheat.Tool && f.Title == cheat.Title {
			return nil // Already favorited
		}
	}
	favs = append(favs, FavoriteEntry{
		Tool:    cheat.Tool,
		Title:   cheat.Title,
		Command: cheat.Command,
	})
	return p.saveJSON("favorites.json", favs)
}

// RemoveFavorite removes a cheat from favorites.
func (p *Persistence) RemoveFavorite(cheat *model.Cheat) error {
	favs, _ := p.LoadFavorites()
	filtered := make([]FavoriteEntry, 0, len(favs))
	for _, f := range favs {
		if f.Tool != cheat.Tool || f.Title != cheat.Title {
			filtered = append(filtered, f)
		}
	}
	return p.saveJSON("favorites.json", filtered)
}

// LoadFavorites loads all favorited cheats.
func (p *Persistence) LoadFavorites() ([]FavoriteEntry, error) {
	var favs []FavoriteEntry
	err := p.loadJSON("favorites.json", &favs)
	return favs, err
}

// IsFavorite checks if a cheat is in favorites.
func (p *Persistence) IsFavorite(cheat *model.Cheat) bool {
	favs, _ := p.LoadFavorites()
	for _, f := range favs {
		if f.Tool == cheat.Tool && f.Title == cheat.Title {
			return true
		}
	}
	return false
}

// =============================================================================
// Argument History
// =============================================================================

// ArgHistoryEntry stores a single argument name-value pair.
type ArgHistoryEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SaveArgHistory saves argument history for later recall.
func (p *Persistence) SaveArgHistory(argName, argValue string) error {
	entries, _ := p.LoadArgHistory(argName)
	// Deduplicate by value
	for i, e := range entries {
		if e.Value == argValue {
			// Move to front
			entries = append(entries[:i], entries[i+1:]...)
			break
		}
	}
	entries = append([]ArgHistoryEntry{{Name: argName, Value: argValue}}, entries...)
	if len(entries) > 20 {
		entries = entries[:20]
	}
	return p.saveJSON("arg_history_"+sanitize(argName)+".json", entries)
}

// LoadArgHistory loads argument history for a given argument name.
func (p *Persistence) LoadArgHistory(argName string) ([]ArgHistoryEntry, error) {
	var entries []ArgHistoryEntry
	err := p.loadJSON("arg_history_"+sanitize(argName)+".json", &entries)
	return entries, err
}

// =============================================================================
// Internal Helpers
// =============================================================================

func (p *Persistence) saveJSON(filename string, data interface{}) error {
	path := filepath.Join(p.configDir, filename)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("ERROR: Failed to marshal %s: %v", filename, err)
		return fmt.Errorf("marshal error: %w", err)
	}
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		log.Printf("ERROR: Failed to write %s: %v", filename, err)
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func (p *Persistence) loadJSON(filename string, target interface{}) error {
	path := filepath.Join(p.configDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist = empty
		}
		return err
	}
	return json.Unmarshal(data, target)
}

// sanitize removes unsafe characters from filenames.
func sanitize(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return "default"
	}
	return string(result)
}
