package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/stackctl/internal/stackctl"
)

type dashTab int

const (
	dashTabOverview dashTab = iota
	dashTabEnv
	dashTabDetail
)

type containerInfo struct {
	Service string
	State   string
	Health  string
	CPU     string
	Mem     string
	Ports   string
}

type envStatus struct {
	Name       string
	Containers []containerInfo
	Status     string // OK, DEGRADED, NOT DEPLOYED
}

type refreshMsg struct {
	envStatuses []envStatus
}

type tickMsg time.Time

func StartDashboard(env string) error {
	m := newDashModel(env)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

type dashModel struct {
	focusEnv    string
	activeTab   dashTab
	envStatuses []envStatus
	envCursor   int
	rowCursor   int
	detailModel *dashDetailModel
	width       int
	height      int
}

func newDashModel(env string) dashModel {
	return dashModel{
		focusEnv:    env,
		activeTab:   dashTabOverview,
		detailModel: newDashDetailModel(),
	}
}

func (m dashModel) Init() tea.Cmd {
	return tea.Batch(m.fetchAll(), tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m dashModel) fetchAll() tea.Cmd {
	return func() tea.Msg {
		envs := stackctl.DetectEnvironments()
		if m.focusEnv != "" {
			envs = []string{m.focusEnv}
		}

		var statuses []envStatus
		for _, env := range envs {
			cfg, err := stackctl.LoadEnvConfig(env)
			if err != nil {
				statuses = append(statuses, envStatus{Name: env, Status: "NOT DEPLOYED"})
				continue
			}
			containers := fetchContainers(cfg)
			status := "OK"
			if len(containers) == 0 {
				status = "NOT DEPLOYED"
			} else {
				for _, c := range containers {
					if c.State != "running" {
						status = "DEGRADED"
						break
					}
				}
			}
			statuses = append(statuses, envStatus{
				Name:       env,
				Containers: containers,
				Status:     status,
			})
		}
		return refreshMsg{envStatuses: statuses}
	}
}

type composePS struct {
	Service string `json:"Service"`
	State   string `json:"State"`
	Health  string `json:"Health"`
	Ports   string `json:"Ports"`
}

func fetchContainers(cfg stackctl.EnvConfig) []containerInfo {
	args := stackctl.ComposeBaseArgs(cfg)
	args = append(args, "ps", "--format", "json")
	out, err := stackctl.RunCmdCapture("docker", args...)
	if err != nil {
		return nil
	}

	var containers []containerInfo
	// docker compose ps --format json outputs one JSON object per line
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ps composePS
		if err := json.Unmarshal([]byte(line), &ps); err != nil {
			continue
		}
		containers = append(containers, containerInfo{
			Service: ps.Service,
			State:   ps.State,
			Health:  ps.Health,
			Ports:   ps.Ports,
		})
	}

	// Fetch stats
	statsArgs := stackctl.ComposeBaseArgs(cfg)
	statsArgs = append(statsArgs, "stats", "--no-stream", "--format", "json")
	statsOut, err := stackctl.RunCmdCapture("docker", statsArgs...)
	if err == nil {
		type dockerStats struct {
			Name    string `json:"Name"`
			CPUPerc string `json:"CPUPerc"`
			MemPerc string `json:"MemPerc"`
		}
		for _, line := range strings.Split(strings.TrimSpace(statsOut), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var ds dockerStats
			if err := json.Unmarshal([]byte(line), &ds); err != nil {
				continue
			}
			for i := range containers {
				if strings.Contains(ds.Name, containers[i].Service) {
					containers[i].CPU = ds.CPUPerc
					containers[i].Mem = ds.MemPerc
				}
			}
		}
	}

	return containers
}

func (m dashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if isQuit(msg) {
			return m, tea.Quit
		}
		switch {
		case msg.String() == "q" || isEsc(msg):
			if m.activeTab == dashTabDetail {
				m.activeTab = dashTabEnv
				return m, nil
			}
			if m.activeTab == dashTabEnv {
				m.activeTab = dashTabOverview
				return m, nil
			}
			return m, tea.Quit
		case isTab(msg):
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil
		case msg.String() == "1":
			m.activeTab = dashTabOverview
		case msg.String() == "2":
			m.activeTab = dashTabEnv
		case msg.String() == "3":
			m.activeTab = dashTabDetail
		}

		switch m.activeTab {
		case dashTabOverview:
			return m.updateOverview(msg)
		case dashTabEnv:
			return m.updateEnv(msg)
		case dashTabDetail:
			return m.updateDetail(msg)
		}

	case refreshMsg:
		m.envStatuses = msg.envStatuses
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.fetchAll(), tickCmd())
	}

	return m, nil
}

func (m dashModel) updateOverview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case isUp(msg) && m.envCursor > 0:
		m.envCursor--
	case isDown(msg) && m.envCursor < len(m.envStatuses)-1:
		m.envCursor++
	case isEnter(msg):
		if m.envCursor < len(m.envStatuses) {
			m.focusEnv = m.envStatuses[m.envCursor].Name
			m.activeTab = dashTabEnv
			m.rowCursor = 0
		}
	}
	return m, nil
}

func (m dashModel) updateEnv(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	es := m.currentEnvStatus()
	if es == nil {
		return m, nil
	}

	switch {
	case isUp(msg) && m.rowCursor > 0:
		m.rowCursor--
	case isDown(msg) && m.rowCursor < len(es.Containers)-1:
		m.rowCursor++
	case isEnter(msg):
		if m.rowCursor < len(es.Containers) {
			m.detailModel.container = &es.Containers[m.rowCursor]
			m.detailModel.envName = es.Name
			m.activeTab = dashTabDetail
		}
	case msg.String() == "r":
		// Restart selected service
		if m.rowCursor < len(es.Containers) {
			svc := es.Containers[m.rowCursor].Service
			return m, m.restartService(es.Name, svc)
		}
	}
	return m, nil
}

func (m dashModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "l":
		// Open logs with tea.ExecProcess
		if m.detailModel.container != nil {
			return m, m.execLogs(m.detailModel.envName, m.detailModel.container.Service)
		}
	case msg.String() == "x":
		// Open shell with tea.ExecProcess
		if m.detailModel.container != nil {
			return m, m.execShell(m.detailModel.envName, m.detailModel.container.Service)
		}
	}
	return m, nil
}

func (m dashModel) currentEnvStatus() *envStatus {
	for i := range m.envStatuses {
		if m.envStatuses[i].Name == m.focusEnv {
			return &m.envStatuses[i]
		}
	}
	if len(m.envStatuses) > 0 && m.envCursor < len(m.envStatuses) {
		return &m.envStatuses[m.envCursor]
	}
	return nil
}

type restartDoneMsg struct{ err error }

func (m dashModel) restartService(env, service string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := stackctl.LoadEnvConfig(env)
		if err != nil {
			return restartDoneMsg{err: err}
		}
		args := stackctl.ComposeBaseArgs(cfg)
		args = append(args, "restart", service)
		_, err = stackctl.RunCmdCapture("docker", args...)
		return restartDoneMsg{err: err}
	}
}

func (m dashModel) execLogs(env, service string) tea.Cmd {
	cfg, err := stackctl.LoadEnvConfig(env)
	if err != nil {
		return nil
	}
	args := stackctl.ComposeBaseArgs(cfg)
	args = append(args, "logs", "-f", service)
	allArgs := append([]string{}, args...)
	c := tea.ExecProcess(execCmd("docker", allArgs...), func(err error) tea.Msg {
		return restartDoneMsg{err: err}
	})
	return c
}

func (m dashModel) execShell(env, service string) tea.Cmd {
	cfg, err := stackctl.LoadEnvConfig(env)
	if err != nil {
		return nil
	}
	args := stackctl.ComposeBaseArgs(cfg)
	args = append(args, "exec", service, "sh")
	allArgs := append([]string{}, args...)
	c := tea.ExecProcess(execCmd("docker", allArgs...), func(err error) tea.Msg {
		return restartDoneMsg{err: err}
	})
	return c
}

func (m dashModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("stackctl Dashboard"))
	b.WriteString("\n")

	// Tabs
	tabs := []string{"Overview", "Environment", "Detail"}
	for i, tab := range tabs {
		if dashTab(i) == m.activeTab {
			b.WriteString(activeTabStyle.Render(tab))
		} else {
			b.WriteString(inactiveTabStyle.Render(tab))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	switch m.activeTab {
	case dashTabOverview:
		b.WriteString(m.viewOverview())
	case dashTabEnv:
		b.WriteString(m.viewEnv())
	case dashTabDetail:
		b.WriteString(m.viewDetail())
	}

	b.WriteString(helpStyle.Render("\n  tab/1-3: switch tabs  j/k: navigate  enter: select  r: restart  l: logs  x: shell  q: quit"))
	return b.String()
}

func (m dashModel) viewOverview() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("  Environments"))
	b.WriteString("\n\n")

	if len(m.envStatuses) == 0 {
		b.WriteString(mutedStyle.Render("  No environments detected. Run 'stackctl init' first."))
		b.WriteString("\n")
		return b.String()
	}

	// Header
	b.WriteString(fmt.Sprintf("  %s%-12s %-14s %-8s%s\n",
		"  ",
		tableHeaderStyle.Render("ENV"),
		tableHeaderStyle.Render("CONTAINERS"),
		tableHeaderStyle.Render("STATUS"),
		""))

	for i, es := range m.envStatuses {
		prefix := "  "
		if i == m.envCursor {
			prefix = cursorChar
		}

		statusStyle := statusRunning
		switch es.Status {
		case "DEGRADED":
			statusStyle = statusUnknown
		case "NOT DEPLOYED":
			statusStyle = statusStopped
		}

		b.WriteString(fmt.Sprintf("  %s %-12s %-14s %s\n",
			prefix,
			normalStyle.Render(es.Name),
			mutedStyle.Render(fmt.Sprintf("%d", len(es.Containers))),
			statusStyle.Render(es.Status)))
	}
	return b.String()
}

func (m dashModel) viewEnv() string {
	var b strings.Builder

	es := m.currentEnvStatus()
	if es == nil {
		b.WriteString(mutedStyle.Render("  Select an environment from Overview tab."))
		return b.String()
	}

	b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %s — Containers", es.Name)))
	b.WriteString("\n\n")

	if len(es.Containers) == 0 {
		b.WriteString(mutedStyle.Render("  No containers running."))
		b.WriteString("\n")
		return b.String()
	}

	// Header
	b.WriteString(fmt.Sprintf("     %-20s %-12s %-10s %-8s %-8s %s\n",
		tableHeaderStyle.Render("SERVICE"),
		tableHeaderStyle.Render("STATE"),
		tableHeaderStyle.Render("HEALTH"),
		tableHeaderStyle.Render("CPU"),
		tableHeaderStyle.Render("MEM"),
		tableHeaderStyle.Render("PORTS")))

	for i, c := range es.Containers {
		prefix := "  "
		if i == m.rowCursor {
			prefix = cursorChar
		}

		stateStyle := statusRunning
		if c.State != "running" {
			stateStyle = statusStopped
		}

		health := c.Health
		if health == "" {
			health = "-"
		}
		cpu := c.CPU
		if cpu == "" {
			cpu = "-"
		}
		mem := c.Mem
		if mem == "" {
			mem = "-"
		}
		ports := c.Ports
		if ports == "" {
			ports = "-"
		}

		b.WriteString(fmt.Sprintf("  %s %-20s %-12s %-10s %-8s %-8s %s\n",
			prefix,
			normalStyle.Render(c.Service),
			stateStyle.Render(c.State),
			mutedStyle.Render(health),
			mutedStyle.Render(cpu),
			mutedStyle.Render(mem),
			mutedStyle.Render(ports)))
	}
	return b.String()
}

func (m dashModel) viewDetail() string {
	return m.detailModel.View()
}
