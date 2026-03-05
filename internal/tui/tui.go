package tui

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(conn *sql.DB, key []byte) (string, error) {
	m := New(conn, key)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	if finalModel, ok := result.(Model); ok {
		return finalModel.ActivateName(), nil
	}
	return "", nil
}
