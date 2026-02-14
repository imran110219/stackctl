package stackctl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type moduleInfo struct {
	Name        string
	Description string
	Ports       []string
}

var moduleCatalog = map[string]moduleInfo{
	"socket-proxy": {
		Name:        "socket-proxy",
		Description: "Docker socket proxy for safer container API access",
		Ports:       []string{"127.0.0.1:2375"},
	},
	"dozzle": {
		Name:        "dozzle",
		Description: "Container log viewer",
		Ports:       []string{"127.0.0.1:9999"},
	},
	"node-exporter": {
		Name:        "node-exporter",
		Description: "Host metrics exporter",
		Ports:       []string{"127.0.0.1:9100"},
	},
	"prometheus": {
		Name:        "prometheus",
		Description: "Metrics scraping and storage",
		Ports:       []string{"127.0.0.1:9090"},
	},
	"alertmanager": {
		Name:        "alertmanager",
		Description: "Alert routing",
		Ports:       []string{"127.0.0.1:9093"},
	},
	"grafana": {
		Name:        "grafana",
		Description: "Dashboards",
		Ports:       []string{"127.0.0.1:3000"},
	},
	"loki": {
		Name:        "loki",
		Description: "Log aggregation",
		Ports:       []string{"127.0.0.1:3100"},
	},
	"jaeger": {
		Name:        "jaeger",
		Description: "Distributed tracing",
		Ports:       []string{"127.0.0.1:16686", "127.0.0.1:4317", "127.0.0.1:4318"},
	},
	"kuma": {
		Name:        "kuma",
		Description: "Uptime Kuma monitoring",
		Ports:       []string{"127.0.0.1:3001"},
	},
	"certbot": {
		Name:        "certbot",
		Description: "Optional certificate management helper",
		Ports:       []string{},
	},
	"backup": {
		Name:        "backup",
		Description: "Backup sidecar tools and hooks",
		Ports:       []string{},
	},
}

var moduleDependencies = map[string][]string{
	"dozzle": {"socket-proxy"},
}

type enabledConfig struct {
	Modules []string `yaml:"modules"`
}

func loadEnabledModules(cfg envConfig) ([]string, error) {
	enabled, err := loadEnabled(cfg)
	if err != nil {
		return nil, err
	}
	if len(enabled.Modules) == 0 {
		return []string{}, nil
	}

	mods := make([]string, 0, len(enabled.Modules))
	for _, m := range enabled.Modules {
		if _, ok := moduleCatalog[m]; ok {
			mods = append(mods, m)
		}
	}
	mods = addModuleDependencies(mods)
	sort.Strings(mods)
	return mods, nil
}

func loadEnabled(cfg envConfig) (enabledConfig, error) {
	path := filepath.Join(cfg.EnvDir, "enabled.yml")
	b, err := os.ReadFile(path)
	if err != nil {
		return enabledConfig{}, err
	}
	var conf enabledConfig
	if err := yaml.Unmarshal(b, &conf); err != nil {
		return enabledConfig{}, err
	}
	return conf, nil
}

func writeEnabled(cfg envConfig, conf enabledConfig) error {
	path := filepath.Join(cfg.EnvDir, "enabled.yml")
	out, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o640)
}

func addModuleDependencies(modules []string) []string {
	set := map[string]bool{}
	for _, m := range modules {
		set[m] = true
	}
	for _, m := range modules {
		for _, dep := range moduleDependencies[m] {
			set[dep] = true
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func sortedModuleNames() []string {
	names := make([]string, 0, len(moduleCatalog))
	for name := range moduleCatalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedModulePorts(name string) string {
	m, ok := moduleCatalog[name]
	if !ok || len(m.Ports) == 0 {
		return "-"
	}
	return strings.Join(m.Ports, ",")
}
