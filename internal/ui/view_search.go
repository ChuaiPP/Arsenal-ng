// Package ui provides the terminal user interface for arsenal-ng.
//
// This file contains the search view rendering functions including the main
// search interface, cheat list display, command hints for special commands,
// and all related rendering helpers.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/halilkirazkaya/arsenal-ng/internal/config"
	"github.com/halilkirazkaya/arsenal-ng/internal/model"
)

// =============================================================================
// Search View
// =============================================================================

func (m App) viewSearch() string {
	var b strings.Builder
	width := m.effectiveWidth()

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Filter status bar
	if m.filterActive {
		b.WriteString(helpStyle.Render("[Filter ON] Ctrl+F to toggle | Ctrl+E: favorites | Ctrl+H: history"))
		b.WriteString("\n")
	}
	if m.showFavorites {
		b.WriteString(successStyle.Render("[Favorites] Showing favorites only | Ctrl+E to toggle"))
		b.WriteString("\n")
	}
	if m.showHistory {
		b.WriteString(successStyle.Render("[History] Showing execution history | Ctrl+H to toggle"))
		b.WriteString("\n")
	}

	// Status message
	if m.statusMsg != "" {
		style := successStyle
		if m.statusIsError {
			style = errorStyle
		}
		b.WriteString(style.Render(m.statusMsg))
		b.WriteString("\n\n")
	}

	// Info box for selected item
	if len(m.filtered) > 0 && m.cursor >= 0 && m.cursor < len(m.filtered) {
		b.WriteString(m.renderInfoBox(m.filtered[m.cursor], width))
		b.WriteString("\n\n")
	}

	// Search input
	b.WriteString(promptStyle.Render("❯ "))
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	// Content: cheat list or command hints
	query := m.searchInput.Value()
	if !isSpecialCommand(query) {
		b.WriteString(m.renderCheatList(query))
	} else {
		b.WriteString(m.renderCommandHints(query))
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m App) renderInfoBox(cheat *model.Cheat, width int) string {
	var content strings.Builder

	content.WriteString(titleStyle.Render(cheat.Title))
	if cheat.Desc != "" {
		content.WriteString("\n")
		// Use wordWrap instead of truncate for desc to allow multi-line wrapping
		wrappedDesc := wordWrap(cheat.Desc, width-6)
		content.WriteString(descStyle.Render(wrappedDesc))
	}
	content.WriteString("\n")
	content.WriteString(syntaxHighlight(truncate(cheat.Command, width-6)))

	return infoBoxStyle.Width(width - 4).Render(content.String())
}

func (m App) renderCheatList(query string) string {
	var b strings.Builder

	maxVisible := m.maxVisibleItems()
	end := m.offset + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	matchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.ColorAccent)).
		Bold(true)

	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderCheatRow(i, m.filtered[i], query, matchStyle))
	}

	return b.String()
}

func (m App) renderCheatRow(index int, cheat *model.Cheat, query string, matchStyle lipgloss.Style) string {
	// Cursor indicator
	cursor := "  "
	if index == m.cursor {
		cursor = cursorStyle.Render("▸ ")
	}

	// Tool name
	toolText := truncate(cheat.Tool, 12)
	tool := fuzzyHighlight(fmt.Sprintf("%-12s", toolText), query, toolStyle, matchStyle)

	// Title
	titleText := truncate(cheat.Title, 40)
	var title string
	if index == m.cursor {
		title = selectedStyle.Render(fmt.Sprintf("%-40s", titleText))
	} else {
		title = fuzzyHighlight(fmt.Sprintf("%-40s", titleText), query, lipgloss.NewStyle(), matchStyle)
	}

	// Tags
	tags := renderColoredTags(cheat.Tags)

	return fmt.Sprintf("%s%s %s %s\n", cursor, tool, title, tags)
}

// =============================================================================
// Command Hints (for special commands)
// =============================================================================

func (m App) renderCommandHints(query string) string {
	var b strings.Builder
	q := strings.ToLower(strings.TrimSpace(query))

	switch {
	case q == "set" || q == "set ":
		b.WriteString(m.renderSetHints())

	case strings.HasPrefix(q, "set ") && !strings.Contains(q, "="):
		partial := strings.TrimPrefix(q, "set ")
		b.WriteString(m.renderSetSuggestions(partial))

	case strings.HasPrefix(q, "set ") && strings.Contains(q, "="):
		b.WriteString(successStyle.Render("✓ Press Enter to set variable"))

	case q == "unset" || q == "unset ":
		b.WriteString(m.renderUnsetHints())

	case q == "variables" || q == "variables ":
		b.WriteString(successStyle.Render("✓ Press Enter to show all variables"))

	case q == "tools":
		b.WriteString(successStyle.Render("✓ Press Enter to show all available tools"))

	case q == "help" || q == "help ":
		b.WriteString(successStyle.Render("✓ Press Enter to show help menu"))

	default:
		b.WriteString(helpStyle.Render("Press Enter to execute command..."))
	}

	return b.String()
}

func (m App) renderSetHints() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("💡 Common variables you can set:"))
	b.WriteString("\n\n")

	for _, v := range commonVariables {
		varName := argNameStyle.Render(fmt.Sprintf("  set %-10s", v.Name+"="))
		desc := helpStyle.Render(v.Desc)
		b.WriteString(fmt.Sprintf("%s %s\n", varName, desc))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Type the full command and press Enter"))

	return b.String()
}

func (m App) renderSetSuggestions(partial string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("💡 Suggestions:"))
	b.WriteString("\n\n")

	found := false
	for _, v := range commonVariables {
		if strings.HasPrefix(v.Name, partial) {
			varName := argNameStyle.Render(fmt.Sprintf("  set %s=", v.Name))
			desc := helpStyle.Render(v.Desc)
			b.WriteString(fmt.Sprintf("%s %s\n", varName, desc))
			found = true
		}
	}

	if !found {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  set %s=<value>  (custom variable)\n", partial)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Add '=' and value, then press Enter"))

	return b.String()
}

func (m App) renderUnsetHints() string {
	var b strings.Builder

	vars := m.globals.All()
	if len(vars) == 0 {
		b.WriteString(titleStyle.Render("  ⚠ No variables to unset"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  Set variables first:"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("    %s\n", argNameStyle.Render("set key=value")))
		b.WriteString(fmt.Sprintf("    %s\n", argNameStyle.Render("set ip=10.10.10.10")))
		b.WriteString(fmt.Sprintf("    %s\n", argNameStyle.Render("set domain=target.com")))
	} else {
		b.WriteString(titleStyle.Render("  💡 Variables you can unset:"))
		b.WriteString("\n\n")
		for name, value := range vars {
			varName := argNameStyle.Render(fmt.Sprintf("unset %-12s", name))
			val := helpStyle.Render(fmt.Sprintf("(current: %s)", value))
			b.WriteString(fmt.Sprintf("    %s  %s\n", varName, val))
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  Type variable name and press Enter"))
	}

	return b.String()
}



// viewShowFavorites renders the favorites view.
func (m App) viewShowFavorites() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("Favorites"))
	b.WriteString("\n\n")

	favs := m.globals.All() // placeholder
    _ = favs
	var favCheats []*model.Cheat
	for _, cheat := range m.cheats {
		if cheat.IsFavorite {
			favCheats = append(favCheats, cheat)
		}
	}

	if len(favCheats) == 0 {
		b.WriteString(helpStyle.Render("No favorites yet. Select a command and press Ctrl+D to favorite it."))
	} else {
		matchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(config.ColorAccent)).Bold(true)
		for i, cheat := range favCheats {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("> ")
			}
			toolText := truncate(cheat.Tool, 12)
			titleText := truncate(cheat.Title, 40)
			tags := renderColoredTags(cheat.Tags)
			b.WriteString(fmt.Sprintf("%s%s %-40s %s\n", cursor,
				fuzzyHighlight(fmt.Sprintf("%-12s", toolText), "", toolStyle, matchStyle),
				titleText, tags))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press ESC to go back | Ctrl+D to unfavorite"))
	return b.String()
}

// viewShowHistory renders the history view.
func (m App) viewShowHistory() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("Execution History"))
	b.WriteString("\n\n")

	if len(m.history) == 0 {
		b.WriteString(helpStyle.Render("No history yet. Commands you execute will appear here."))
	} else {
		for i, rec := range m.history {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("> ")
			}
			toolText := truncate(rec.Tool, 15)
			cmdText := truncate(rec.Command, 60)
			b.WriteString(fmt.Sprintf("%s%-15s %s\n", cursor,
				toolStyle.Render(toolText), helpStyle.Render(cmdText)))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press ESC to go back"))
	return b.String()
}
