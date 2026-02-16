package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var logo = `
 ███████╗████████╗ █████╗  ██████╗██╗  ██╗ ██████╗████████╗██╗
 ██╔════╝╚══██╔══╝██╔══██╗██╔════╝██║ ██╔╝██╔════╝╚══██╔══╝██║
 ███████╗   ██║   ███████║██║     █████╔╝ ██║        ██║   ██║
 ╚════██║   ██║   ██╔══██║██║     ██╔═██╗ ██║        ██║   ██║
 ███████║   ██║   ██║  ██║╚██████╗██║  ██╗╚██████╗   ██║   ███████╗
 ╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝   ╚═╝   ╚══════╝
`

type menuItem struct {
	label string
	desc  string
}

type welcomeModel struct {
	cursor int
	items  []menuItem
}

func newWelcomeModel() *welcomeModel {
	return &welcomeModel{
		items: []menuItem{
			{label: "New Setup", desc: "Initialize a new environment from scratch"},
			{label: "Add Environment", desc: "Add another env (qa/prod) to existing stack"},
			{label: "Exit", desc: "Quit the setup wizard"},
		},
	}
}

func (m *welcomeModel) Init() tea.Cmd {
	return nil
}

func (m *welcomeModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isUp(msg) {
			if m.cursor > 0 {
				m.cursor--
			}
		}
		if isDown(msg) {
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		}
		if isEnter(msg) {
			switch m.cursor {
			case 0, 1:
				return m, func() tea.Msg {
					return navigateMsg{to: screenEnvSelect}
				}
			case 2:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *welcomeModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(logo))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Interactive Setup Wizard"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("  %s %s\n", cursorChar, selectedStyle.Render(item.label)))
			b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render(item.desc)))
		} else {
			b.WriteString(fmt.Sprintf("    %s\n", normalStyle.Render(item.label)))
			b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render(item.desc)))
		}
	}

	b.WriteString(helpStyle.Render("\n  up/down: navigate  enter: select  ?: help  ctrl+c: quit"))
	return b.String()
}
