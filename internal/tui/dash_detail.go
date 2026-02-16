package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type dashDetailModel struct {
	container *containerInfo
	envName   string
}

func newDashDetailModel() *dashDetailModel {
	return &dashDetailModel{}
}

func (m *dashDetailModel) View() string {
	if m.container == nil {
		return mutedStyle.Render("  Select a container from the Environment tab.")
	}

	var b strings.Builder

	b.WriteString(subtitleStyle.Render(fmt.Sprintf("  Container: %s", m.container.Service)))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Service:  %s\n", normalStyle.Render(m.container.Service)))
	b.WriteString(fmt.Sprintf("  State:    %s\n", stateStyleFor(m.container.State).Render(m.container.State)))

	health := m.container.Health
	if health == "" {
		health = "-"
	}
	b.WriteString(fmt.Sprintf("  Health:   %s\n", mutedStyle.Render(health)))

	cpu := m.container.CPU
	if cpu == "" {
		cpu = "-"
	}
	b.WriteString(fmt.Sprintf("  CPU:      %s\n", mutedStyle.Render(cpu)))

	mem := m.container.Mem
	if mem == "" {
		mem = "-"
	}
	b.WriteString(fmt.Sprintf("  Memory:   %s\n", mutedStyle.Render(mem)))

	ports := m.container.Ports
	if ports == "" {
		ports = "-"
	}
	b.WriteString(fmt.Sprintf("  Ports:    %s\n", mutedStyle.Render(ports)))

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  l: view logs  x: open shell  esc: back"))

	return b.String()
}

func stateStyleFor(state string) lipgloss.Style {
	if state == "running" {
		return statusRunning
	}
	return statusStopped
}

func execCmd(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
