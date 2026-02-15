package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type stepStatus int

const (
	stepPending stepStatus = iota
	stepRunning
	stepDone
	stepFailed
)

type progressStep struct {
	label  string
	status stepStatus
	err    error
}

type stepDoneMsg struct {
	index int
	err   error
}

type progressModel struct {
	state   *wizardState
	steps   []progressStep
	spinner spinner.Model
	current int
	done    bool
	errMsg  string
}

func newProgressModel(state *wizardState) *progressModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &progressModel{
		state:   state,
		spinner: sp,
		steps: []progressStep{
			{label: "Initializing environment"},
			{label: "Enabling modules"},
			{label: "Applying configuration"},
		},
	}
}

func (m *progressModel) Init() tea.Cmd {
	// Reset state for fresh run
	m.current = 0
	m.done = false
	m.errMsg = ""
	for i := range m.steps {
		m.steps[i].status = stepPending
		m.steps[i].err = nil
	}
	m.steps[0].status = stepRunning

	return tea.Batch(m.spinner.Tick, m.runStep(0))
}

func (m *progressModel) runStep(index int) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch index {
		case 0:
			err = m.doInit()
		case 1:
			err = m.doEnable()
		case 2:
			err = m.doApply()
		}
		return stepDoneMsg{index: index, err: err}
	}
}

func captureOutput(fn func() error) (string, error) {
	oldOut, oldErr := os.Stdout, os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout, os.Stderr = w, w
	err := fn()
	w.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

func (m *progressModel) doInit() error {
	cfg, err := stackctl.LoadEnvConfig(m.state.env)
	if err != nil {
		return err
	}
	cfg.Domain = m.state.domain
	cfg.Email = m.state.email

	_, err = captureOutput(func() error {
		return stackctl.RunInit(cfg)
	})
	return err
}

func (m *progressModel) doEnable() error {
	if len(m.state.modules) == 0 {
		return nil
	}

	cfg, err := stackctl.LoadEnvConfig(m.state.env)
	if err != nil {
		return err
	}

	modules := make([]string, len(m.state.modules))
	copy(modules, m.state.modules)
	modules = stackctl.AddModuleDependencies(modules)
	sort.Strings(modules)

	conf := stackctl.EnabledConfig{Modules: modules}
	return stackctl.WriteEnabled(cfg, conf)
}

func (m *progressModel) doApply() error {
	_, err := captureOutput(func() error {
		return stackctl.Run([]string{"apply", "--env", m.state.env})
	})
	return err
}

func (m *progressModel) Update(msg tea.Msg) (screenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case stepDoneMsg:
		m.steps[msg.index].status = stepDone
		if msg.err != nil {
			m.steps[msg.index].status = stepFailed
			m.steps[msg.index].err = msg.err
			m.errMsg = msg.err.Error()
			m.done = true
			return m, nil
		}

		next := msg.index + 1
		if next >= len(m.steps) {
			m.done = true
			return m, func() tea.Msg { return navigateMsg{to: screenComplete} }
		}
		m.current = next
		m.steps[next].status = stepRunning
		return m, m.runStep(next)

	case tea.KeyMsg:
		if m.done && m.errMsg != "" {
			if isEnter(msg) || isEsc(msg) {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *progressModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Setting Up"))
	b.WriteString("\n\n")

	for _, step := range m.steps {
		var icon string
		switch step.status {
		case stepPending:
			icon = mutedStyle.Render("  ")
		case stepRunning:
			icon = m.spinner.View()
		case stepDone:
			icon = successStyle.Render("OK")
		case stepFailed:
			icon = errorStyle.Render("XX")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", icon, normalStyle.Render(step.label)))
	}

	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("  Error: " + m.errMsg))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("\n  press enter or esc to exit"))
	}

	return b.String()
}
