package tui

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(conn *sql.DB, key []byte) error {
	m := New(conn, key)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
