package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	state  *wizardState
	cursor int
}

func newConfirmModel(state *wizardState) *confirmModel {
	return &confirmModel{state: state}
}

func (m *confirmModel) Init() tea.Cmd {
	m.cursor = 0
	return nil
}

func (m *confirmModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) {
			return m, func() tea.Msg { return navigateMsg{to: screenModuleSelect} }
		}
		if isLeft(msg) && m.cursor > 0 {
			m.cursor--
		}
		if isRight(msg) && m.cursor < 2 {
			m.cursor++
		}
		if isUp(msg) && m.cursor > 0 {
			m.cursor--
		}
		if isDown(msg) && m.cursor < 2 {
			m.cursor++
		}
		if isEnter(msg) {
			switch m.cursor {
			case 0: // Confirm
				return m, func() tea.Msg { return navigateMsg{to: screenPreflight} }
			case 1: // Back
				return m, func() tea.Msg { return navigateMsg{to: screenModuleSelect} }
			case 2: // Cancel
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *confirmModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm Setup"))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("  Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Environment:  %s\n", selectedStyle.Render(m.state.env)))
	b.WriteString(fmt.Sprintf("  Domain:       %s\n", selectedStyle.Render(m.state.domain)))
	b.WriteString(fmt.Sprintf("  Email:        %s\n", selectedStyle.Render(m.state.email)))

	if len(m.state.modules) > 0 {
		b.WriteString(fmt.Sprintf("  Modules:      %s\n", selectedStyle.Render(strings.Join(m.state.modules, ", "))))
	} else {
		b.WriteString(fmt.Sprintf("  Modules:      %s\n", mutedStyle.Render("(none)")))
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  Equivalent CLI Commands"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl init --env %s --domain %s --email %s",
		m.state.env, m.state.domain, m.state.email)))
	b.WriteString("\n")
	for _, mod := range m.state.modules {
		b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl enable %s --env %s", mod, m.state.env)))
		b.WriteString("\n")
	}
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  $ stackctl apply --env %s", m.state.env)))
	b.WriteString("\n\n")

	buttons := []string{"Confirm", "Back", "Cancel"}
	for i, btn := range buttons {
		if i == m.cursor {
			b.WriteString("  " + borderStyle.Render(selectedStyle.Render(btn)))
		} else {
			b.WriteString("  " + normalStyle.Render("["+btn+"]"))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("\n  left/right: navigate  enter: select  esc: back"))
	return b.String()
}

