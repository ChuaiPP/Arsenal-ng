// Package ui provides the terminal user interface for arsenal-ng.
//
// This file contains the argument input view rendering functions. It displays
// a form where users can fill in command arguments before execution.
package ui

import (
	"fmt"
	"strings"

	"github.com/halilkirazkaya/arsenal-ng/internal/model"
)

// =============================================================================
// Args View
// =============================================================================

func (m App) viewArgs() string {
	var b strings.Builder
	width := m.effectiveWidth()

	// Command preview
	preview := model.BuildCommand(m.selectedCheat.Command, m.args)
	previewBox := infoBoxStyle.Width(width - 4).Render(
		titleStyle.Render("Command Preview") + "\n" + syntaxHighlight(preview),
	)
	b.WriteString(previewBox)
	b.WriteString("\n\n")

	// Arguments list
	b.WriteString(titleStyle.Render("Arguments:"))
	b.WriteString("\n\n")

	for i, arg := range m.args {
		b.WriteString(m.renderArgRow(i, arg))
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("tab/↓: next │ shift+tab/↑: prev │ enter: confirm │ esc: back"))

	return b.String()
}

func (m App) renderArgRow(index int, arg model.Argument) string {
	cursor := "  "
	if index == m.argCursor {
		cursor = cursorStyle.Render("▸ ")
	}

	name := argNameStyle.Render(fmt.Sprintf("%-15s", arg.Name))
	input := m.argInputs[index].View()

	// Show available options as hints
	row := fmt.Sprintf("%s%s = %s", cursor, name, input)
	if len(m.selectedCheat.Options) > 0 {
		for _, opt := range m.selectedCheat.Options {
			if opt.Arg == arg.Name {
				row += "  " + helpStyle.Render(fmt.Sprintf("[%s]", opt.Value))
			}
		}
	}
	return row + "\n"
}

