package tui

import (
	"fmt"
	"strings"

	"github.com/example/stackctl/internal/stackctl"
)

type modulesDetailModel struct {
	module  string
	cfg     stackctl.EnvConfig
	enabled map[string]bool
}

func newModulesDetailModel() *modulesDetailModel {
	return &modulesDetailModel{}
}

func (m *modulesDetailModel) View() string {
	if m.module == "" {
		return ""
	}

	info, ok := stackctl.ModuleCatalog[m.module]
	if !ok {
		return ""
	}

	var b strings.Builder

	b.WriteString(borderStyle.Render(subtitleStyle.Render(info.Name)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n", normalStyle.Render(info.Description)))
	b.WriteString(fmt.Sprintf("  Category: %s\n", mutedStyle.Render(info.Category)))

	// Ports
	if len(info.Ports) > 0 {
		b.WriteString(fmt.Sprintf("  Ports:    %s\n", mutedStyle.Render(strings.Join(info.Ports, ", "))))
	} else {
		b.WriteString(fmt.Sprintf("  Ports:    %s\n", mutedStyle.Render("none")))
	}

	// Dependencies
	if deps, ok := stackctl.ModuleDependencies[m.module]; ok && len(deps) > 0 {
		b.WriteString(fmt.Sprintf("  Depends:  %s\n", mutedStyle.Render(strings.Join(deps, ", "))))
	}

	// Reverse dependencies
	var rdeps []string
	for mod, deps := range stackctl.ModuleDependencies {
		for _, dep := range deps {
			if dep == m.module {
				rdeps = append(rdeps, mod)
			}
		}
	}
	if len(rdeps) > 0 {
		b.WriteString(fmt.Sprintf("  Needed by: %s\n", mutedStyle.Render(strings.Join(rdeps, ", "))))
	}

	// Status
	if m.enabled[m.module] {
		if stackctl.ComposeServiceRunning(m.cfg, m.module) {
			b.WriteString(fmt.Sprintf("  Status:   %s\n", statusRunning.Render("running")))
		} else {
			b.WriteString(fmt.Sprintf("  Status:   %s\n", statusStopped.Render("stopped")))
		}
	} else {
		b.WriteString(fmt.Sprintf("  Status:   %s\n", mutedStyle.Render("disabled")))
	}

	return b.String()
}
