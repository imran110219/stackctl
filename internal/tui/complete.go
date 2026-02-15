package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type completeModel struct {
	state *wizardState
}

func newCompleteModel(state *wizardState) *completeModel {
	return &completeModel{state: state}
}

func (m *completeModel) Init() tea.Cmd {
	return nil
}

func (m *completeModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEnter(msg) || msg.String() == "q" || isEsc(msg) {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *completeModel) View() string {
	var b strings.Builder

	b.WriteString(successStyle.Render("  Setup Complete!"))
	b.WriteString("\n\n")

	cfg, _ := stackctl.LoadEnvConfig(m.state.env)
	b.WriteString(fmt.Sprintf("  Environment:  %s\n", selectedStyle.Render(m.state.env)))
	b.WriteString(fmt.Sprintf("  Path:         %s\n", normalStyle.Render(cfg.EnvDir)))
	b.WriteString(fmt.Sprintf("  Domain:       %s\n", normalStyle.Render(m.state.domain)))

	if len(m.state.modules) > 0 {
		b.WriteString(fmt.Sprintf("  Modules:      %s\n", normalStyle.Render(strings.Join(m.state.modules, ", "))))
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  Next Steps"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl status --env %s      # check status", m.state.env)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl enable <mod> --env %s # add more modules", m.state.env)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl apply --env %s       # re-apply changes", m.state.env)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl doctor                # verify system")))
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("\n  press q or enter to exit"))
	return b.String()
}
