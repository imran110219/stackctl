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
	screenProgress
	screenComplete
)

type navigateMsg struct {
	to screen
}

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
		screenProgress:     newProgressModel(state),
		screenComplete:     newCompleteModel(state),
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

	case navigateMsg:
		m.current = msg.to
		s := m.screens[m.current]
		initCmd := s.Init()
		return m, initCmd
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

	if m.current != screenProgress && m.current != screenComplete {
		step := int(m.current)
		total := int(screenComplete)
		if step > 0 && step < total {
			progress := mutedStyle.Render(fmt.Sprintf("Step %d of %d", step, total-1))
			content = content + "\n" + progress
		}
	}

	return content
}
