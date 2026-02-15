package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type envOption struct {
	value string
	label string
	desc  string
}

type envSelectModel struct {
	state   *wizardState
	cursor  int
	options []envOption
}

func newEnvSelectModel(state *wizardState) *envSelectModel {
	return &envSelectModel{
		state: state,
		options: []envOption{
			{value: "dev", label: "dev", desc: "Development environment"},
			{value: "qa", label: "qa", desc: "QA / staging environment"},
			{value: "prod", label: "prod", desc: "Production environment"},
		},
	}
}

func (m *envSelectModel) Init() tea.Cmd {
	// Restore cursor position if going back
	for i, opt := range m.options {
		if opt.value == m.state.env {
			m.cursor = i
			break
		}
	}
	return nil
}

func (m *envSelectModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) {
			return m, func() tea.Msg { return navigateMsg{to: screenWelcome} }
		}
		if isUp(msg) && m.cursor > 0 {
			m.cursor--
		}
		if isDown(msg) && m.cursor < len(m.options)-1 {
			m.cursor++
		}
		if isEnter(msg) {
			m.state.env = m.options[m.cursor].value
			return m, func() tea.Msg { return navigateMsg{to: screenDomainInput} }
		}
	}
	return m, nil
}

func (m *envSelectModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Environment"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Choose the target environment for this setup."))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		radio := radioOff
		label := normalStyle.Render(opt.label)
		if i == m.cursor {
			radio = radioOn
			label = selectedStyle.Render(opt.label)
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", radio, label))
		b.WriteString(fmt.Sprintf("      %s\n", mutedStyle.Render(opt.desc)))
	}

	b.WriteString(helpStyle.Render("\n  up/down: navigate  enter: select  esc: back"))
	return b.String()
}
