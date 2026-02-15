package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type emailInputModel struct {
	state  *wizardState
	input  textinput.Model
	errMsg string
}

func newEmailInputModel(state *wizardState) *emailInputModel {
	ti := textinput.New()
	ti.Placeholder = "admin@example.com"
	ti.CharLimit = 254
	ti.Width = 40

	return &emailInputModel{
		state: state,
		input: ti,
	}
}

func (m *emailInputModel) Init() tea.Cmd {
	if m.state.email != "" {
		m.input.SetValue(m.state.email)
	}
	m.input.Focus()
	return textinput.Blink
}

func (m *emailInputModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) {
			return m, func() tea.Msg { return navigateMsg{to: screenDomainInput} }
		}
		if isEnter(msg) {
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				val = "admin@example.com"
			}
			if !emailRegex.MatchString(val) {
				m.errMsg = "Invalid email format"
				return m, nil
			}
			m.errMsg = ""
			m.state.email = val
			return m, func() tea.Msg { return navigateMsg{to: screenModuleSelect} }
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *emailInputModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Email"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Enter the admin/ops email for notifications and certificates."))
	b.WriteString("\n\n")
	b.WriteString("  " + m.input.View())
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString("\n  " + errorStyle.Render(m.errMsg))
	}

	b.WriteString(helpStyle.Render("\n  enter: confirm  esc: back"))
	return b.String()
}
