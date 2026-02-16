package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type validationResult struct {
	key     string
	ok      bool
	message string
}

type configValidateModel struct {
	cfg     stackctl.EnvConfig
	vars    map[string]string
	results []validationResult
}

func newConfigValidateModel(cfg stackctl.EnvConfig) *configValidateModel {
	return &configValidateModel{cfg: cfg}
}

func (m *configValidateModel) Init() tea.Cmd {
	m.results = m.validate()
	return nil
}

func (m *configValidateModel) validate() []validationResult {
	var results []validationResult

	// Check required fields
	required := []string{"DOMAIN", "ADMIN_EMAIL"}
	for _, key := range required {
		val, exists := m.vars[key]
		if !exists || val == "" {
			results = append(results, validationResult{key: key, ok: false, message: "required but missing"})
		} else {
			results = append(results, validationResult{key: key, ok: true, message: "set"})
		}
	}

	// Check placeholder values
	placeholders := map[string]string{
		"DOMAIN":      "example.com",
		"ADMIN_EMAIL": "admin@example.com",
	}
	for key, placeholder := range placeholders {
		if m.vars[key] == placeholder {
			results = append(results, validationResult{key: key, ok: false, message: fmt.Sprintf("still using placeholder '%s'", placeholder)})
		}
	}

	// Check password lengths
	for key := range secretKeys {
		val, exists := m.vars[key]
		if !exists {
			continue
		}
		if len(val) < 8 {
			results = append(results, validationResult{key: key, ok: false, message: "password too short (< 8 chars)"})
		} else {
			results = append(results, validationResult{key: key, ok: true, message: fmt.Sprintf("%d chars", len(val))})
		}
	}

	return results
}

func (m *configValidateModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) || isEnter(msg) || msg.String() == "q" {
			return m, func() tea.Msg {
				return configNavigateMsg{to: configScreenEditor}
			}
		}
	}
	return m, nil
}

func (m *configValidateModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Configuration Validation"))
	b.WriteString("\n\n")

	allOK := true
	for _, r := range m.results {
		icon := successStyle.Render("OK")
		if !r.ok {
			icon = warningStyle.Render("!!")
			allOK = false
		}
		b.WriteString(fmt.Sprintf("  %s %-30s %s\n", icon, normalStyle.Render(r.key), mutedStyle.Render(r.message)))
	}

	b.WriteString("\n")
	if allOK {
		b.WriteString(successStyle.Render("  All checks passed!"))
	} else {
		b.WriteString(warningStyle.Render("  Some issues found. Review above."))
	}

	b.WriteString(helpStyle.Render("\n\n  enter/esc: back to editor"))
	return b.String()
}
