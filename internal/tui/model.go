package tui

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/feto/umgebung/internal/db"
)

var (
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)

type Model struct {
	conn     *sql.DB
	key      []byte
	sets     []db.EnvSet
	cursor   int
	vars     []db.EnvVar
	width    int
	height   int
	err      error
	quitting bool
}

type setsLoadedMsg struct{ sets []db.EnvSet }
type varsLoadedMsg struct{ vars []db.EnvVar }
type errMsg struct{ err error }

func New(conn *sql.DB, key []byte) Model {
	return Model{conn: conn, key: key}
}

func (m Model) Init() tea.Cmd {
	return m.loadSets()
}

func (m Model) loadSets() tea.Cmd {
	return func() tea.Msg {
		sets, err := db.ListEnvSets(m.conn)
		if err != nil {
			return errMsg{err}
		}
		return setsLoadedMsg{sets}
	}
}

func (m Model) loadVars() tea.Cmd {
	if len(m.sets) == 0 {
		return nil
	}
	name := m.sets[m.cursor].Name
	return func() tea.Msg {
		vars, err := db.GetEnvVars(m.conn, name)
		if err != nil {
			return errMsg{err}
		}
		return varsLoadedMsg{vars}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case setsLoadedMsg:
		m.sets = msg.sets
		m.cursor = 0
		return m, m.loadVars()

	case varsLoadedMsg:
		m.vars = msg.vars

	case errMsg:
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadVars()
			}
		case "down", "j":
			if m.cursor < len(m.sets)-1 {
				m.cursor++
				return m, m.loadVars()
			}
		case "d":
			if len(m.sets) > 0 {
				name := m.sets[m.cursor].Name
				db.DeleteEnvSet(m.conn, name)
				return m, m.loadSets()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}
	if m.quitting {
		return ""
	}

	leftWidth := m.width / 3
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := m.width - leftWidth - 3
	if rightWidth < 20 {
		rightWidth = 20
	}

	// Left pane: env set list
	var left strings.Builder
	left.WriteString(headerStyle.Render("EnvSets") + "\n")
	for i, s := range m.sets {
		line := s.Name
		if i == m.cursor {
			line = "> " + selectedStyle.Render(line)
		} else {
			line = "  " + line
		}
		left.WriteString(line + "\n")
	}
	if len(m.sets) == 0 {
		left.WriteString(dimStyle.Render("  (empty)") + "\n")
	}
	left.WriteString("\n" + helpStyle.Render("[n]ew [d]el [q]uit"))

	// Right pane: var preview
	activeSet := "(none)"
	if len(m.sets) > 0 {
		activeSet = m.sets[m.cursor].Name
	}
	var right strings.Builder
	right.WriteString(headerStyle.Render("Preview: "+activeSet) + "\n")
	for _, v := range m.vars {
		val := strings.Repeat("*", 16)
		line := fmt.Sprintf("  %-20s = %s", v.Key, dimStyle.Render(val))
		right.WriteString(line + "\n")
	}
	if len(m.vars) == 0 {
		right.WriteString(dimStyle.Render("  (no variables)") + "\n")
	}
	right.WriteString("\n" + helpStyle.Render("[e]dit [u]p [enter] activate"))

	leftPane := lipgloss.NewStyle().Width(leftWidth).Render(left.String())
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Render(strings.Repeat("│\n", m.height))
	rightPane := lipgloss.NewStyle().Width(rightWidth).Render(right.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, divider, rightPane)
}
