package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type configScreen int

const (
	configScreenEditor configScreen = iota
	configScreenValidate
	configScreenRestart
)

func StartConfigWizard(env string) error {
	if env == "" {
		envs := stackctl.DetectEnvironments()
		if len(envs) == 0 {
			return fmt.Errorf("no environments found; run 'stackctl init' first")
		}
		env = envs[0]
	}

	cfg, err := stackctl.LoadEnvConfig(env)
	if err != nil {
		return err
	}

	m := newConfigRootModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

type configRootModel struct {
	cfg     stackctl.EnvConfig
	current configScreen
	editor  *configEditorModel
	valid   *configValidateModel
	restart *configRestartModel
}

func newConfigRootModel(cfg stackctl.EnvConfig) configRootModel {
	editor := newConfigEditorModel(cfg)
	return configRootModel{
		cfg:     cfg,
		current: configScreenEditor,
		editor:  editor,
		valid:   newConfigValidateModel(cfg),
		restart: newConfigRestartModel(cfg),
	}
}

type configNavigateMsg struct {
	to          configScreen
	changedKeys []string
}

func (m configRootModel) Init() tea.Cmd {
	return m.editor.Init()
}

func (m configRootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isQuit(msg) {
			return m, tea.Quit
		}

	case configNavigateMsg:
		m.current = msg.to
		if msg.to == configScreenRestart {
			m.restart.changedKeys = msg.changedKeys
		}
		switch m.current {
		case configScreenEditor:
			return m, m.editor.Init()
		case configScreenValidate:
			m.valid.vars = m.editor.vars
			return m, m.valid.Init()
		case configScreenRestart:
			return m, m.restart.Init()
		}
		return m, nil
	}

	switch m.current {
	case configScreenEditor:
		newEditor, cmd := m.editor.Update(msg)
		m.editor = newEditor.(*configEditorModel)
		return m, cmd
	case configScreenValidate:
		newValid, cmd := m.valid.Update(msg)
		m.valid = newValid.(*configValidateModel)
		return m, cmd
	case configScreenRestart:
		newRestart, cmd := m.restart.Update(msg)
		m.restart = newRestart.(*configRestartModel)
		return m, cmd
	}

	return m, nil
}

func (m configRootModel) View() string {
	switch m.current {
	case configScreenEditor:
		return m.editor.View()
	case configScreenValidate:
		return m.valid.View()
	case configScreenRestart:
		return m.restart.View()
	}
	return ""
}
