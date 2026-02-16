package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type helpReturnMsg struct{}

type helpModel struct{}

func newHelpModel() *helpModel {
	return &helpModel{}
}

func (m *helpModel) Init() tea.Cmd {
	return nil
}

func (m *helpModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) || msg.String() == "?" || isEnter(msg) || msg.String() == "q" {
			return m, func() tea.Msg { return helpReturnMsg{} }
		}
	}
	return m, nil
}

func (m *helpModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{
			title: "Global",
			keys: []struct{ key, desc string }{
				{"ctrl+c", "Quit immediately"},
				{"?", "Toggle this help screen"},
			},
		},
		{
			title: "Navigation",
			keys: []struct{ key, desc string }{
				{"up / k", "Move up"},
				{"down / j", "Move down"},
				{"left / h", "Move left / previous option"},
				{"right / l", "Move right / next option"},
				{"enter", "Confirm / select"},
				{"esc", "Go back / cancel"},
				{"tab", "Switch tabs (dashboard)"},
			},
		},
		{
			title: "Setup Wizard",
			keys: []struct{ key, desc string }{
				{"space", "Toggle module selection"},
				{"enter", "Confirm and proceed"},
				{"esc", "Go back to previous step"},
			},
		},
		{
			title: "Module Manager",
			keys: []struct{ key, desc string }{
				{"space", "Toggle module enabled/disabled"},
				{"/", "Search/filter modules"},
				{"d", "Toggle detail pane"},
				{"s", "Save changes"},
				{"a", "Save + apply"},
				{"q", "Quit"},
			},
		},
		{
			title: "Dashboard",
			keys: []struct{ key, desc string }{
				{"tab / 1-3", "Switch tabs"},
				{"r", "Restart selected service"},
				{"l", "View logs for selected service"},
				{"x", "Open shell in selected service"},
				{"q", "Quit"},
			},
		},
		{
			title: "Config Editor",
			keys: []struct{ key, desc string }{
				{"enter", "Edit selected variable"},
				{"u", "Unmask/mask secret value"},
				{"g", "Generate secure password"},
				{"v", "Validate configuration"},
				{"s", "Save changes"},
				{"q", "Quit"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(categoryStyle.Render("  " + section.title))
		b.WriteString("\n")
		for _, k := range section.keys {
			b.WriteString(subtitleStyle.Render("    "+k.key) + "  " + mutedStyle.Render(k.desc))
			b.WriteString("\n")
		}
	}

	b.WriteString(helpStyle.Render("\n  press ?, esc, or enter to close"))
	return b.String()
}
