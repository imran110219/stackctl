package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*\.)+[a-zA-Z]{2,}$`)

type domainInputModel struct {
	state    *wizardState
	input    textinput.Model
	errMsg   string
}

func newDomainInputModel(state *wizardState) *domainInputModel {
	ti := textinput.New()
	ti.Placeholder = "example.com"
	ti.CharLimit = 253
	ti.Width = 40

	return &domainInputModel{
		state: state,
		input: ti,
	}
}

func (m *domainInputModel) Init() tea.Cmd {
	if m.state.domain != "" {
		m.input.SetValue(m.state.domain)
	}
	m.input.Focus()
	return textinput.Blink
}

func (m *domainInputModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) {
			return m, func() tea.Msg { return navigateMsg{to: screenEnvSelect} }
		}
		if isEnter(msg) {
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				val = "example.com"
			}
			if !domainRegex.MatchString(val) {
				m.errMsg = "Invalid domain format"
				return m, nil
			}
			m.errMsg = ""
			m.state.domain = val
			return m, func() tea.Msg { return navigateMsg{to: screenEmailInput} }
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *domainInputModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Domain"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Enter the base domain for this environment."))
	b.WriteString("\n\n")
	b.WriteString("  " + m.input.View())
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString("\n  " + errorStyle.Render(m.errMsg))
	}

	b.WriteString(helpStyle.Render("\n  enter: confirm  esc: back"))
	return b.String()
}
