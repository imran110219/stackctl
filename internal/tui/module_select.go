package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type moduleRow struct {
	name       string
	isCategory bool
	category   string
}

type moduleSelectModel struct {
	state    *wizardState
	rows     []moduleRow
	cursor   int
	selected map[string]bool
	depMsg   string
}

func newModuleSelectModel(state *wizardState) *moduleSelectModel {
	m := &moduleSelectModel{
		state:    state,
		selected: map[string]bool{},
	}
	m.buildRows()
	return m
}

func (m *moduleSelectModel) buildRows() {
	categories := []string{"Observability", "Infrastructure", "Utilities"}
	grouped := map[string][]string{}
	for name, info := range stackctl.ModuleCatalog {
		grouped[info.Category] = append(grouped[info.Category], name)
	}

	m.rows = nil
	for _, cat := range categories {
		names := grouped[cat]
		if len(names) == 0 {
			continue
		}
		sort.Strings(names)
		m.rows = append(m.rows, moduleRow{isCategory: true, category: cat})
		for _, name := range names {
			m.rows = append(m.rows, moduleRow{name: name})
		}
	}
}

func (m *moduleSelectModel) Init() tea.Cmd {
	// Restore selections from state
	for _, mod := range m.state.modules {
		m.selected[mod] = true
	}
	// Skip past first category header
	if len(m.rows) > 0 && m.rows[0].isCategory {
		m.cursor = 1
	}
	return nil
}

func (m *moduleSelectModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEsc(msg) {
			return m, func() tea.Msg { return navigateMsg{to: screenEmailInput} }
		}
		if isUp(msg) {
			m.cursor--
			for m.cursor >= 0 && m.rows[m.cursor].isCategory {
				m.cursor--
			}
			if m.cursor < 0 {
				// Find first non-category
				for i, r := range m.rows {
					if !r.isCategory {
						m.cursor = i
						break
					}
				}
			}
		}
		if isDown(msg) {
			m.cursor++
			for m.cursor < len(m.rows) && m.rows[m.cursor].isCategory {
				m.cursor++
			}
			if m.cursor >= len(m.rows) {
				m.cursor = len(m.rows) - 1
				for m.cursor >= 0 && m.rows[m.cursor].isCategory {
					m.cursor--
				}
			}
		}
		if isSpace(msg) {
			row := m.rows[m.cursor]
			if !row.isCategory {
				m.depMsg = ""
				if m.selected[row.name] {
					delete(m.selected, row.name)
				} else {
					m.selected[row.name] = true
					// Auto-resolve dependencies
					if deps, ok := stackctl.ModuleDependencies[row.name]; ok {
						for _, dep := range deps {
							if !m.selected[dep] {
								m.selected[dep] = true
								m.depMsg = fmt.Sprintf("auto-enabled %s (required by %s)", dep, row.name)
							}
						}
					}
				}
			}
		}
		if isEnter(msg) {
			m.state.modules = nil
			for name := range m.selected {
				m.state.modules = append(m.state.modules, name)
			}
			sort.Strings(m.state.modules)
			return m, func() tea.Msg { return navigateMsg{to: screenConfirm} }
		}
	}
	return m, nil
}

func (m *moduleSelectModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Modules"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Choose which modules to enable. Dependencies are auto-resolved."))
	b.WriteString("\n")

	for i, row := range m.rows {
		if row.isCategory {
			b.WriteString(categoryStyle.Render("  " + row.category))
			b.WriteString("\n")
			continue
		}

		info := stackctl.ModuleCatalog[row.name]
		check := checkOff
		if m.selected[row.name] {
			check = checkOn
		}

		prefix := "  "
		label := normalStyle.Render(info.Name)
		if i == m.cursor {
			prefix = cursorChar
			label = selectedStyle.Render(info.Name)
		}

		ports := "-"
		if len(info.Ports) > 0 {
			ports = strings.Join(info.Ports, ", ")
		}

		b.WriteString(fmt.Sprintf("  %s %s %s  %s",
			prefix, check, label,
			mutedStyle.Render(info.Description)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("          %s\n", mutedStyle.Render("ports: "+ports)))
	}

	if m.depMsg != "" {
		b.WriteString("\n  " + warningStyle.Render(m.depMsg))
	}

	b.WriteString(helpStyle.Render("\n  up/down: navigate  space: toggle  enter: confirm  esc: back"))
	return b.String()
}
