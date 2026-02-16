package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type modulesListModel struct {
	cfg             stackctl.EnvConfig
	rows            []moduleRow
	cursor          int
	enabled         map[string]bool
	dirty           bool
	searching       bool
	searchInput     textinput.Model
	searchFilter    string
	filteredRows    []int
	showDetail      bool
	detailModel     *modulesDetailModel
	showQuitWarning bool
	statusMsg       string
}

func newModulesListModel(cfg stackctl.EnvConfig) *modulesListModel {
	ti := textinput.New()
	ti.Placeholder = "search modules..."
	ti.CharLimit = 50
	ti.Width = 30

	m := &modulesListModel{
		cfg:         cfg,
		enabled:     map[string]bool{},
		searchInput: ti,
	}
	m.buildRows()
	m.detailModel = newModulesDetailModel()
	return m
}

func (m *modulesListModel) buildRows() {
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

func (m *modulesListModel) Init() tea.Cmd {
	// Load current enabled modules
	enabled, err := stackctl.LoadEnabled(m.cfg)
	if err == nil {
		for _, mod := range enabled.Modules {
			m.enabled[mod] = true
		}
	}
	// Position cursor on first non-category row
	for i, r := range m.rows {
		if !r.isCategory {
			m.cursor = i
			break
		}
	}
	m.updateFilter()
	return nil
}

func (m *modulesListModel) updateFilter() {
	m.filteredRows = nil
	filter := strings.ToLower(m.searchFilter)
	for i, row := range m.rows {
		if row.isCategory {
			m.filteredRows = append(m.filteredRows, i)
			continue
		}
		if filter == "" {
			m.filteredRows = append(m.filteredRows, i)
			continue
		}
		info := stackctl.ModuleCatalog[row.name]
		if strings.Contains(strings.ToLower(row.name), filter) ||
			strings.Contains(strings.ToLower(info.Description), filter) {
			m.filteredRows = append(m.filteredRows, i)
		}
	}
}

func (m *modulesListModel) visibleIdx(cursor int) int {
	for vi, ri := range m.filteredRows {
		if ri == cursor {
			return vi
		}
	}
	return -1
}

func (m *modulesListModel) nextVisible(from int, dir int) int {
	vi := m.visibleIdx(from)
	if vi < 0 {
		if len(m.filteredRows) > 0 {
			return m.filteredRows[0]
		}
		return from
	}
	for {
		vi += dir
		if vi < 0 || vi >= len(m.filteredRows) {
			return from
		}
		ri := m.filteredRows[vi]
		if !m.rows[ri].isCategory {
			return ri
		}
	}
}

func (m *modulesListModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	if m.searching {
		return m.updateSearch(msg)
	}

	switch msg := msg.(type) {
	case saveMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Save error: %v", msg.err)
		} else {
			m.dirty = false
			m.statusMsg = "Saved!"
		}
		return m, nil

	case applyMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Apply error: %v", msg.err)
		} else {
			m.statusMsg = "Applied successfully!"
		}
		return m, nil

	case tea.KeyMsg:
		if m.showQuitWarning {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
			default:
				m.showQuitWarning = false
				return m, nil
			}
		}

		switch {
		case isUp(msg):
			m.cursor = m.nextVisible(m.cursor, -1)
		case isDown(msg):
			m.cursor = m.nextVisible(m.cursor, 1)
		case isSpace(msg):
			m.toggleModule()
		case msg.String() == "d":
			m.showDetail = !m.showDetail
		case isSlash(msg):
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case msg.String() == "s":
			return m, m.save()
		case msg.String() == "a":
			return m, tea.Batch(m.save(), m.apply())
		case msg.String() == "q" || isEsc(msg):
			if m.dirty {
				m.showQuitWarning = true
				return m, nil
			}
			return m, tea.Quit
		}
	}

	if m.showDetail && m.cursor < len(m.rows) && !m.rows[m.cursor].isCategory {
		m.detailModel.module = m.rows[m.cursor].name
		m.detailModel.cfg = m.cfg
		m.detailModel.enabled = m.enabled
	}

	return m, nil
}

func (m *modulesListModel) updateSearch(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isEnter(msg) || isEsc(msg) {
			m.searching = false
			m.searchInput.Blur()
			if isEsc(msg) {
				m.searchFilter = ""
				m.searchInput.SetValue("")
			} else {
				m.searchFilter = m.searchInput.Value()
			}
			m.updateFilter()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchFilter = m.searchInput.Value()
	m.updateFilter()
	return m, cmd
}

func (m *modulesListModel) toggleModule() {
	if m.cursor >= len(m.rows) || m.rows[m.cursor].isCategory {
		return
	}
	name := m.rows[m.cursor].name
	if m.enabled[name] {
		delete(m.enabled, name)
	} else {
		m.enabled[name] = true
		// Auto-resolve dependencies
		if deps, ok := stackctl.ModuleDependencies[name]; ok {
			for _, dep := range deps {
				if !m.enabled[dep] {
					m.enabled[dep] = true
					m.statusMsg = fmt.Sprintf("auto-enabled %s (required by %s)", dep, name)
				}
			}
		}
	}
	m.dirty = true
}

type saveMsg struct{ err error }
type applyMsg struct{ err error }

func (m *modulesListModel) save() tea.Cmd {
	return func() tea.Msg {
		modules := make([]string, 0, len(m.enabled))
		for name := range m.enabled {
			modules = append(modules, name)
		}
		sort.Strings(modules)
		conf := stackctl.EnabledConfig{Modules: modules}
		err := stackctl.WriteEnabled(m.cfg, conf)
		return saveMsg{err: err}
	}
}

func (m *modulesListModel) apply() tea.Cmd {
	return func() tea.Msg {
		err := stackctl.Run([]string{"apply", "--env", m.cfg.EnvName})
		return applyMsg{err: err}
	}
}

func (m *modulesListModel) View() string {
	var b strings.Builder

	// Header
	title := fmt.Sprintf("Module Manager — %s", m.cfg.EnvName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	if m.dirty {
		b.WriteString(warningStyle.Render("  [unsaved changes]"))
		b.WriteString("\n")
	}

	if m.searching {
		b.WriteString("  " + m.searchInput.View())
		b.WriteString("\n")
	}

	// Module list
	for _, ri := range m.filteredRows {
		row := m.rows[ri]
		if row.isCategory {
			b.WriteString(categoryStyle.Render("  " + row.category))
			b.WriteString("\n")
			continue
		}

		info := stackctl.ModuleCatalog[row.name]
		check := checkOff
		if m.enabled[row.name] {
			check = checkOn
		}

		prefix := "  "
		label := normalStyle.Render(info.Name)
		if ri == m.cursor {
			prefix = cursorChar
			label = selectedStyle.Render(info.Name)
		}

		// Running status
		status := ""
		if m.enabled[row.name] {
			if stackctl.ComposeServiceRunning(m.cfg, row.name) {
				status = statusRunning.Render(" [running]")
			} else {
				status = statusStopped.Render(" [stopped]")
			}
		}

		b.WriteString(fmt.Sprintf("  %s %s %s  %s%s\n",
			prefix, check, label,
			mutedStyle.Render(info.Description), status))
	}

	if m.statusMsg != "" {
		b.WriteString("\n  " + warningStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	if m.showQuitWarning {
		b.WriteString("\n  " + warningStyle.Render("Unsaved changes! Press 'y' to quit or any key to cancel."))
		b.WriteString("\n")
	}

	// Detail pane
	if m.showDetail && m.cursor < len(m.rows) && !m.rows[m.cursor].isCategory {
		b.WriteString("\n")
		b.WriteString(m.detailModel.View())
	}

	b.WriteString(helpStyle.Render("\n  j/k: navigate  space: toggle  /: search  d: detail  s: save  a: apply  q: quit"))
	return b.String()
}
