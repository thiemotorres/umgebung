package tui

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/editor"
)

var (
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)

type Model struct {
	conn         *sql.DB
	key          []byte
	sets         []db.EnvSet
	cursor       int
	vars         []db.EnvVar
	width        int
	height       int
	err          error
	quitting     bool
	mode         string // "browse" | "input"
	inputBuf     string // name being typed in input mode
	activateName string // set when user presses enter to activate
}

type setsLoadedMsg struct{ sets []db.EnvSet }
type varsLoadedMsg struct{ vars []db.EnvVar }
type errMsg struct{ err error }

func New(conn *sql.DB, key []byte) Model {
	return Model{conn: conn, key: key, mode: "browse"}
}

// Mode returns the current mode ("browse" or "input").
func (m Model) Mode() string { return m.mode }

// InputBuf returns the current input buffer.
func (m Model) InputBuf() string { return m.inputBuf }

// ActivateName returns the name set when the user presses enter to activate.
func (m Model) ActivateName() string { return m.activateName }

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

	case error:
		m.err = msg

	case nil:
		return m, m.loadSets()

	case tea.KeyMsg:
		if m.mode == "input" {
			switch msg.Type {
			case tea.KeyEsc:
				m.mode = "browse"
				m.inputBuf = ""
			case tea.KeyBackspace:
				if len(m.inputBuf) > 0 {
					m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
				}
			case tea.KeyEnter:
				if m.inputBuf != "" {
					name := m.inputBuf
					m.mode = "browse"
					m.inputBuf = ""
					return m, m.openEditorForNew(name)
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.inputBuf += string(msg.Runes)
				}
			}
			return m, nil
		}
		// browse mode
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
		case "n":
			m.mode = "input"
			m.inputBuf = ""
		case "d":
			if len(m.sets) > 0 {
				name := m.sets[m.cursor].Name
				db.DeleteEnvSet(m.conn, name)
				return m, m.loadSets()
			}
		case "e":
			if len(m.sets) > 0 {
				return m, m.openEditorForEdit()
			}
		case "enter":
			if len(m.sets) > 0 {
				m.activateName = m.sets[m.cursor].Name
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) openEditorForNew(name string) tea.Cmd {
	editorBin := os.Getenv("EDITOR")
	if editorBin == "" {
		editorBin = "vi"
	}
	f, err := os.CreateTemp("", "umgebung-*.env")
	if err != nil {
		return func() tea.Msg { return errMsg{err} }
	}
	tmpPath := f.Name()
	f.WriteString("# umgebung - one KEY=VALUE per line\n")
	f.Close()

	return tea.ExecProcess(exec.Command(editorBin, tmpPath), func(err error) tea.Msg {
		defer os.Remove(tmpPath)
		if err != nil {
			return errMsg{err}
		}
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return errMsg{err}
		}
		pairs, err := editor.ParseLines(string(content))
		if err != nil {
			return errMsg{err}
		}
		var vars []db.EnvVar
		for _, p := range pairs {
			enc, err := crypto.Encrypt(m.key, []byte(p.Value))
			if err != nil {
				return errMsg{err}
			}
			vars = append(vars, db.EnvVar{Key: p.Key, Value: enc})
		}
		if len(vars) > 0 {
			if err := db.CreateEnvSet(m.conn, name, vars); err != nil {
				return errMsg{err}
			}
		}
		return nil // trigger reload via nil case in Update
	})
}

func (m Model) openEditorForEdit() tea.Cmd {
	if len(m.sets) == 0 {
		return nil
	}
	name := m.sets[m.cursor].Name
	editorBin := os.Getenv("EDITOR")
	if editorBin == "" {
		editorBin = "vi"
	}

	vars, err := db.GetEnvVars(m.conn, name)
	if err != nil {
		return func() tea.Msg { return errMsg{err} }
	}

	var initial []editor.EnvPair
	for _, v := range vars {
		plaintext, err := crypto.Decrypt(m.key, v.Value)
		if err != nil {
			return func() tea.Msg { return errMsg{err} }
		}
		initial = append(initial, editor.EnvPair{Key: v.Key, Value: string(plaintext)})
	}

	f, err := os.CreateTemp("", "umgebung-*.env")
	if err != nil {
		return func() tea.Msg { return errMsg{err} }
	}
	tmpPath := f.Name()
	f.WriteString("# umgebung - one KEY=VALUE per line\n")
	f.WriteString(editor.FormatLines(initial))
	f.Close()

	return tea.ExecProcess(exec.Command(editorBin, tmpPath), func(err error) tea.Msg {
		defer os.Remove(tmpPath)
		if err != nil {
			return errMsg{err}
		}
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return errMsg{err}
		}
		pairs, err := editor.ParseLines(string(content))
		if err != nil {
			return errMsg{err}
		}
		var updatedVars []db.EnvVar
		for _, p := range pairs {
			enc, err := crypto.Encrypt(m.key, []byte(p.Value))
			if err != nil {
				return errMsg{err}
			}
			updatedVars = append(updatedVars, db.EnvVar{Key: p.Key, Value: enc})
		}
		return db.UpdateEnvSet(m.conn, name, updatedVars)
	})
}

func (m Model) View() string {
	if m.mode == "input" {
		return fmt.Sprintf("New env set name: %s_\n\n(press enter to open editor, esc to cancel)", m.inputBuf)
	}

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
