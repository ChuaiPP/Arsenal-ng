package ui

import (
	"github.com/halilkirazkaya/arsenal-ng/internal/loader"
	tea "github.com/charmbracelet/bubbletea"
)

// updateShowFavorites handles input in the favorites view.
func (m App) updateShowFavorites(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case keyEsc, keyQuit, keyEnter:
			m.showFavorites = false
			m.state = stateSearch
			m.filtered = loader.Search(m.cheats, m.searchInput.Value())
			m.cursor = -1
			m.searchInput.Focus()
			return m, nil
		}
	}
	return m, nil
}
