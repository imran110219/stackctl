package stackctl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type ModuleInfo struct {
	Name        string
	Description string
	Ports       []string
	Category    string
}

var ModuleCatalog = map[string]ModuleInfo{
	"socket-proxy": {
		Name:        "socket-proxy",
		Description: "Docker socket proxy for safer container API access",
		Ports:       []string{"127.0.0.1:2375"},
		Category:    "Infrastructure",
	},
	"dozzle": {
		Name:        "dozzle",
		Description: "Container log viewer",
		Ports:       []string{"127.0.0.1:9999"},
		Category:    "Observability",
	},
	"node-exporter": {
		Name:        "node-exporter",
		Description: "Host metrics exporter",
		Ports:       []string{"127.0.0.1:9100"},
		Category:    "Observability",
	},
	"prometheus": {
		Name:        "prometheus",
		Description: "Metrics scraping and storage",
		Ports:       []string{"127.0.0.1:9090"},
		Category:    "Observability",
	},
	"alertmanager": {
		Name:        "alertmanager",
		Description: "Alert routing",
		Ports:       []string{"127.0.0.1:9093"},
		Category:    "Observability",
	},
	"grafana": {
		Name:        "grafana",
		Description: "Dashboards",
		Ports:       []string{"127.0.0.1:3000"},
		Category:    "Observability",
	},
	"loki": {
		Name:        "loki",
		Description: "Log aggregation",
		Ports:       []string{"127.0.0.1:3100"},
		Category:    "Observability",
	},
	"jaeger": {
		Name:        "jaeger",
		Description: "Distributed tracing",
		Ports:       []string{"127.0.0.1:16686", "127.0.0.1:4317", "127.0.0.1:4318"},
		Category:    "Observability",
	},
	"kuma": {
		Name:        "kuma",
		Description: "Uptime Kuma monitoring",
		Ports:       []string{"127.0.0.1:3001"},
		Category:    "Infrastructure",
	},
	"certbot": {
		Name:        "certbot",
		Description: "Optional certificate management helper",
		Ports:       []string{},
		Category:    "Infrastructure",
	},
	"backup": {
		Name:        "backup",
		Description: "Backup sidecar tools and hooks",
		Ports:       []string{},
		Category:    "Utilities",
	},
}

var ModuleDependencies = map[string][]string{
	"dozzle": {"socket-proxy"},
}

type EnabledConfig struct {
	Modules []string `yaml:"modules"`
}

func LoadEnabledModules(cfg EnvConfig) ([]string, error) {
	enabled, err := LoadEnabled(cfg)
	if err != nil {
		return nil, err
	}
	if len(enabled.Modules) == 0 {
		return []string{}, nil
	}

	mods := make([]string, 0, len(enabled.Modules))
	for _, m := range enabled.Modules {
		if _, ok := ModuleCatalog[m]; ok {
			mods = append(mods, m)
		}
	}
	mods = AddModuleDependencies(mods)
	sort.Strings(mods)
	return mods, nil
}

func LoadEnabled(cfg EnvConfig) (EnabledConfig, error) {
	path := filepath.Join(cfg.EnvDir, "enabled.yml")
	b, err := os.ReadFile(path)
	if err != nil {
		return EnabledConfig{}, err
	}
	var conf EnabledConfig
	if err := yaml.Unmarshal(b, &conf); err != nil {
		return EnabledConfig{}, err
	}
	return conf, nil
}

func WriteEnabled(cfg EnvConfig, conf EnabledConfig) error {
	path := filepath.Join(cfg.EnvDir, "enabled.yml")
	out, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o640)
}

func AddModuleDependencies(modules []string) []string {
	set := map[string]bool{}
	for _, m := range modules {
		set[m] = true
	}
	for _, m := range modules {
		for _, dep := range ModuleDependencies[m] {
			set[dep] = true
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func SortedModuleNames() []string {
	names := make([]string, 0, len(ModuleCatalog))
	for name := range ModuleCatalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedModulePorts(name string) string {
	m, ok := ModuleCatalog[name]
	if !ok || len(m.Ports) == 0 {
		return "-"
	}
	return strings.Join(m.Ports, ",")
}
