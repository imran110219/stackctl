package tui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

var secretKeys = map[string]bool{
	"POSTGRES_PASSWORD":   true,
	"MYSQL_ROOT_PASSWORD": true,
	"RESTIC_PASSWORD":     true,
	"KC_DB_PASSWORD":      true,
	"KEYCLOAK_ADMIN_PASSWORD": true,
	"SECRET_KEY":          true,
	"JWT_SECRET":          true,
}

var keyGroups = []struct {
	name string
	keys []string
}{
	{"Core", []string{"DOMAIN", "ADMIN_EMAIL", "ENV_NAME"}},
	{"Databases", []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "MYSQL_ROOT_PASSWORD", "KC_DB_PASSWORD"}},
	{"Security", []string{"SECRET_KEY", "JWT_SECRET", "KEYCLOAK_ADMIN_PASSWORD"}},
	{"Backup", []string{"RESTIC_REPOSITORY", "RESTIC_PASSWORD"}},
}

type configEditorModel struct {
	cfg        stackctl.EnvConfig
	vars       map[string]string
	keys       []string
	cursor     int
	editing    bool
	editInput  textinput.Model
	unmasked   map[string]bool
	dirty      bool
	statusMsg  string
	origVars   map[string]string
}

func newConfigEditorModel(cfg stackctl.EnvConfig) *configEditorModel {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 50

	return &configEditorModel{
		cfg:       cfg,
		vars:      map[string]string{},
		unmasked:  map[string]bool{},
		editInput: ti,
		origVars:  map[string]string{},
	}
}

func (m *configEditorModel) Init() tea.Cmd {
	envPath := filepath.Join(m.cfg.EnvDir, ".env")
	vars, err := stackctl.ReadDotEnv(envPath)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Error loading .env: %v", err)
		return nil
	}
	m.vars = vars
	// Save originals for change detection
	for k, v := range vars {
		m.origVars[k] = v
	}
	m.buildKeys()
	return nil
}

func (m *configEditorModel) buildKeys() {
	seen := map[string]bool{}
	m.keys = nil

	// Add grouped keys first
	for _, g := range keyGroups {
		for _, k := range g.keys {
			if _, exists := m.vars[k]; exists {
				m.keys = append(m.keys, k)
				seen[k] = true
			}
		}
	}

	// Add remaining keys sorted
	var remaining []string
	for k := range m.vars {
		if !seen[k] {
			remaining = append(remaining, k)
		}
	}
	sort.Strings(remaining)
	m.keys = append(m.keys, remaining...)
}

func (m *configEditorModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	if m.editing {
		return m.updateEdit(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case isUp(msg) && m.cursor > 0:
			m.cursor--
		case isDown(msg) && m.cursor < len(m.keys)-1:
			m.cursor++
		case isEnter(msg):
			if m.cursor < len(m.keys) {
				m.editing = true
				key := m.keys[m.cursor]
				m.editInput.SetValue(m.vars[key])
				m.editInput.Focus()
				return m, textinput.Blink
			}
		case msg.String() == "u":
			// Toggle unmask for current key
			if m.cursor < len(m.keys) {
				key := m.keys[m.cursor]
				if secretKeys[key] {
					m.unmasked[key] = !m.unmasked[key]
				}
			}
		case msg.String() == "g":
			// Generate password
			if m.cursor < len(m.keys) {
				key := m.keys[m.cursor]
				if secretKeys[key] {
					pw, err := generatePassword(32)
					if err == nil {
						m.vars[key] = pw
						m.dirty = true
						m.statusMsg = fmt.Sprintf("Generated password for %s", key)
					}
				}
			}
		case msg.String() == "v":
			return m, func() tea.Msg {
				return configNavigateMsg{to: configScreenValidate}
			}
		case msg.String() == "s":
			return m, m.save()
		case msg.String() == "q" || isEsc(msg):
			return m, tea.Quit
		}
	case saveMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Save error: %v", msg.err)
		} else {
			m.dirty = false
			m.statusMsg = "Saved!"
			// Detect changed keys
			var changed []string
			for k, v := range m.vars {
				if m.origVars[k] != v {
					changed = append(changed, k)
				}
			}
			if len(changed) > 0 {
				return m, func() tea.Msg {
					return configNavigateMsg{to: configScreenRestart, changedKeys: changed}
				}
			}
		}
	}

	return m, nil
}

func (m *configEditorModel) updateEdit(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEnter(msg) {
			key := m.keys[m.cursor]
			m.vars[key] = m.editInput.Value()
			m.editing = false
			m.editInput.Blur()
			m.dirty = true
			return m, nil
		}
		if isEsc(msg) {
			m.editing = false
			m.editInput.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m *configEditorModel) save() tea.Cmd {
	return func() tea.Msg {
		envPath := filepath.Join(m.cfg.EnvDir, ".env")
		err := stackctl.WriteDotEnv(envPath, m.vars)
		return saveMsg{err: err}
	}
}

func (m *configEditorModel) View() string {
	var b strings.Builder

	title := fmt.Sprintf("Configuration Editor — %s", m.cfg.EnvName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	if m.dirty {
		b.WriteString(warningStyle.Render("  [unsaved changes]"))
		b.WriteString("\n")
	}

	currentGroup := ""
	for i, key := range m.keys {
		// Group header
		group := groupForKey(key)
		if group != currentGroup {
			currentGroup = group
			b.WriteString(categoryStyle.Render("  " + group))
			b.WriteString("\n")
		}

		prefix := "  "
		keyStyle := normalStyle
		if i == m.cursor {
			prefix = cursorChar
			keyStyle = selectedStyle
		}

		val := m.vars[key]
		displayVal := val
		if secretKeys[key] && !m.unmasked[key] {
			displayVal = secretStyle.Render("********")
		} else {
			displayVal = normalStyle.Render(val)
		}

		if m.editing && i == m.cursor {
			b.WriteString(fmt.Sprintf("  %s %s = %s\n",
				prefix, keyStyle.Render(key), m.editInput.View()))
		} else {
			b.WriteString(fmt.Sprintf("  %s %s = %s\n",
				prefix, keyStyle.Render(key), displayVal))
		}
	}

	if m.statusMsg != "" {
		b.WriteString("\n  " + successStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("\n  j/k: navigate  enter: edit  u: unmask  g: generate password  v: validate  s: save  q: quit"))
	return b.String()
}

func groupForKey(key string) string {
	for _, g := range keyGroups {
		for _, k := range g.keys {
			if k == key {
				return g.name
			}
		}
	}
	return "Other"
}

func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
