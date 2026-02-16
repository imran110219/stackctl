package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type checksDoneMsg struct {
	results []stackctl.CheckResult
}

type preflightModel struct {
	state    *wizardState
	spinner  spinner.Model
	running  bool
	results  []stackctl.CheckResult
	hasWarn  bool
	cursor   int // 0=Continue, 1=Cancel
}

func newPreflightModel(state *wizardState) *preflightModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &preflightModel{
		state:   state,
		spinner: sp,
	}
}

func (m *preflightModel) Init() tea.Cmd {
	m.running = true
	m.results = nil
	m.hasWarn = false
	m.cursor = 0
	return tea.Batch(m.spinner.Tick, m.runChecks())
}

func (m *preflightModel) runChecks() tea.Cmd {
	return func() tea.Msg {
		results := stackctl.RunChecks()
		return checksDoneMsg{results: results}
	}
}

func (m *preflightModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case checksDoneMsg:
		m.running = false
		m.results = msg.results
		for _, r := range m.results {
			if !r.OK {
				m.hasWarn = true
				break
			}
		}
		if !m.hasWarn {
			// Auto-navigate to progress on all-pass
			return m, func() tea.Msg { return navigateMsg{to: screenProgress} }
		}
		return m, nil

	case tea.KeyMsg:
		if !m.running && m.hasWarn {
			if isLeft(msg) && m.cursor > 0 {
				m.cursor--
			}
			if isRight(msg) && m.cursor < 1 {
				m.cursor++
			}
			if isEnter(msg) {
				if m.cursor == 0 {
					return m, func() tea.Msg { return navigateMsg{to: screenProgress} }
				}
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *preflightModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Pre-flight Checks"))
	b.WriteString("\n\n")

	if m.running {
		b.WriteString(fmt.Sprintf("  %s Running system checks...\n", m.spinner.View()))
		return b.String()
	}

	for _, r := range m.results {
		if r.OK {
			b.WriteString(fmt.Sprintf("  %s %s\n", successStyle.Render("OK"), normalStyle.Render(r.Name)))
		} else {
			b.WriteString(fmt.Sprintf("  %s %s: %s\n",
				warningStyle.Render("!!"),
				normalStyle.Render(r.Name),
				mutedStyle.Render(r.Err.Error())))
		}
	}

	if m.hasWarn {
		b.WriteString("\n")
		b.WriteString(warningStyle.Render("  Some checks have warnings. Continue anyway?"))
		b.WriteString("\n\n")

		buttons := []string{"Continue", "Cancel"}
		for i, btn := range buttons {
			if i == m.cursor {
				b.WriteString("  " + borderStyle.Render(selectedStyle.Render(btn)))
			} else {
				b.WriteString("  " + normalStyle.Render("["+btn+"]"))
			}
			b.WriteString("  ")
		}
		b.WriteString(helpStyle.Render("\n\n  left/right: navigate  enter: select"))
	}

	return b.String()
}
