package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

func StartModuleManager(env string) error {
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

	m := newModulesRootModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

type modulesRootModel struct {
	cfg  stackctl.EnvConfig
	list *modulesListModel
}

func newModulesRootModel(cfg stackctl.EnvConfig) modulesRootModel {
	return modulesRootModel{
		cfg:  cfg,
		list: newModulesListModel(cfg),
	}
}

func (m modulesRootModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m modulesRootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isQuit(msg) {
			if m.list.dirty {
				m.list.showQuitWarning = true
				return m, nil
			}
			return m, tea.Quit
		}
	}

	newList, cmd := m.list.Update(msg)
	m.list = newList.(*modulesListModel)
	return m, cmd
}

func (m modulesRootModel) View() string {
	return m.list.View()
}
