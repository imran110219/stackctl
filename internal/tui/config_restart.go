package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

// Key-to-service mapping for restart detection
var keyServiceMap = map[string][]string{
	"POSTGRES_PASSWORD":       {"postgres"},
	"POSTGRES_USER":           {"postgres"},
	"POSTGRES_DB":             {"postgres"},
	"MYSQL_ROOT_PASSWORD":     {"mariadb"},
	"KC_DB_PASSWORD":          {"keycloak"},
	"KEYCLOAK_ADMIN_PASSWORD": {"keycloak"},
	"DOMAIN":                  {"nginx", "keycloak", "frontend", "backend"},
	"ADMIN_EMAIL":             {"certbot"},
	"SECRET_KEY":              {"backend"},
	"JWT_SECRET":              {"backend"},
}

type configRestartModel struct {
	cfg          stackctl.EnvConfig
	changedKeys  []string
	services     []string
	cursor       int
	applyDone    bool
	applyErr     string
}

func newConfigRestartModel(cfg stackctl.EnvConfig) *configRestartModel {
	return &configRestartModel{cfg: cfg}
}

func (m *configRestartModel) Init() tea.Cmd {
	// Detect affected services from changed keys
	serviceSet := map[string]bool{}
	for _, key := range m.changedKeys {
		if svcs, ok := keyServiceMap[key]; ok {
			for _, s := range svcs {
				serviceSet[s] = true
			}
		}
	}
	m.services = nil
	for s := range serviceSet {
		m.services = append(m.services, s)
	}
	m.cursor = 0
	m.applyDone = false
	m.applyErr = ""
	return nil
}

func (m *configRestartModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.applyDone {
			return m, func() tea.Msg {
				return configNavigateMsg{to: configScreenEditor}
			}
		}

		switch {
		case isLeft(msg) && m.cursor > 0:
			m.cursor--
		case isRight(msg) && m.cursor < 1:
			m.cursor++
		case isEnter(msg):
			if m.cursor == 0 {
				// Apply Now
				return m, m.restartServices()
			}
			// Later
			return m, func() tea.Msg {
				return configNavigateMsg{to: configScreenEditor}
			}
		case isEsc(msg):
			return m, func() tea.Msg {
				return configNavigateMsg{to: configScreenEditor}
			}
		}

	case restartDoneMsg:
		m.applyDone = true
		if msg.err != nil {
			m.applyErr = msg.err.Error()
		}
	}

	return m, nil
}

func (m *configRestartModel) restartServices() tea.Cmd {
	return func() tea.Msg {
		args := stackctl.ComposeBaseArgs(m.cfg)
		args = append(args, "restart")
		args = append(args, m.services...)
		_, err := stackctl.RunCmdCapture("docker", args...)
		return restartDoneMsg{err: err}
	}
}

func (m *configRestartModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Restart Services"))
	b.WriteString("\n\n")

	if len(m.changedKeys) > 0 {
		b.WriteString(subtitleStyle.Render("  Changed variables:"))
		b.WriteString("\n")
		for _, k := range m.changedKeys {
			b.WriteString(fmt.Sprintf("  - %s\n", normalStyle.Render(k)))
		}
		b.WriteString("\n")
	}

	if len(m.services) > 0 {
		b.WriteString(subtitleStyle.Render("  Affected services:"))
		b.WriteString("\n")
		for _, s := range m.services {
			b.WriteString(fmt.Sprintf("  - %s\n", warningStyle.Render(s)))
		}
		b.WriteString("\n")
	} else {
		b.WriteString(mutedStyle.Render("  No service restart needed."))
		b.WriteString("\n\n")
	}

	if m.applyDone {
		if m.applyErr != "" {
			b.WriteString(errorStyle.Render("  Restart error: " + m.applyErr))
		} else {
			b.WriteString(successStyle.Render("  Services restarted successfully!"))
		}
		b.WriteString(helpStyle.Render("\n\n  press any key to continue"))
	} else if len(m.services) > 0 {
		buttons := []string{"Apply Now", "Later"}
		for i, btn := range buttons {
			if i == m.cursor {
				b.WriteString("  " + borderStyle.Render(selectedStyle.Render(btn)))
			} else {
				b.WriteString("  " + normalStyle.Render("["+btn+"]"))
			}
			b.WriteString("  ")
		}
		b.WriteString(helpStyle.Render("\n\n  left/right: navigate  enter: select  esc: back"))
	} else {
		b.WriteString(helpStyle.Render("  press enter or esc to continue"))
	}

	return b.String()
}
