package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenWelcome screen = iota
	screenEnvSelect
	screenDomainInput
	screenEmailInput
	screenModuleSelect
	screenConfirm
	screenPreflight
	screenProgress
	screenComplete
	screenHelp
)

type navigateMsg struct {
	to screen
}

type resetMsg struct{}

type wizardState struct {
	env     string
	domain  string
	email   string
	modules []string
}

type screenModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (screenModel, tea.Cmd)
	View() string
}

type rootModel struct {
	current  screen
	previous screen
	state    *wizardState
	screens  map[screen]screenModel
	width    int
	height   int
	quitting bool
}

func StartWizard() error {
	state := &wizardState{}
	screens := map[screen]screenModel{
		screenWelcome:      newWelcomeModel(),
		screenEnvSelect:    newEnvSelectModel(state),
		screenDomainInput:  newDomainInputModel(state),
		screenEmailInput:   newEmailInputModel(state),
		screenModuleSelect: newModuleSelectModel(state),
		screenConfirm:      newConfirmModel(state),
		screenPreflight:    newPreflightModel(state),
		screenProgress:     newProgressModel(state),
		screenComplete:     newCompleteModel(state),
		screenHelp:         newHelpModel(),
	}

	m := rootModel{
		current: screenWelcome,
		state:   state,
		screens: screens,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m rootModel) Init() tea.Cmd {
	return m.screens[m.current].Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if isQuit(msg) {
			m.quitting = true
			return m, tea.Quit
		}
		// Help overlay accessible via '?' from any non-progress screen
		if msg.String() == "?" && m.current != screenProgress && m.current != screenHelp {
			m.previous = m.current
			m.current = screenHelp
			return m, m.screens[m.current].Init()
		}

	case navigateMsg:
		m.current = msg.to
		s := m.screens[m.current]
		initCmd := s.Init()
		return m, initCmd

	case resetMsg:
		m.state.env = ""
		m.state.domain = ""
		m.state.email = ""
		m.state.modules = nil
		// Recreate module select to clear selections
		m.screens[screenModuleSelect] = newModuleSelectModel(m.state)
		m.current = screenEnvSelect
		s := m.screens[m.current]
		return m, s.Init()

	case helpReturnMsg:
		m.current = m.previous
		return m, nil
	}

	s := m.screens[m.current]
	newScreen, cmd := s.Update(msg)
	m.screens[m.current] = newScreen
	return m, cmd
}

func (m rootModel) View() string {
	if m.quitting {
		return ""
	}

	s := m.screens[m.current]
	content := s.View()

	// Show step indicator for wizard screens (not preflight, progress, complete, help)
	if m.current != screenPreflight && m.current != screenProgress &&
		m.current != screenComplete && m.current != screenHelp {
		step := int(m.current)
		total := int(screenConfirm) // Last "step" screen
		if step > 0 && step <= total {
			progress := mutedStyle.Render(fmt.Sprintf("Step %d of %d", step, total))
			content = content + "\n" + progress
		}
	}

	return content
}
