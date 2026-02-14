package stackctl

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeNginxConfs(cfg envConfig, modules []string) error {
	confDir := filepath.Join(cfg.EnvDir, "nginx", "conf.d")
	if err := ensureDir(confDir, 0o750); err != nil {
		return err
	}

	templates := findTemplatesDir()
	data := cfg.renderData()

	render := func(templateName, targetName string) error {
		inPath := filepath.Join(templates, "nginx", templateName)
		text, err := renderFile(inPath, data)
		if err != nil {
			return fmt.Errorf("render nginx %s: %w", templateName, err)
		}
		return os.WriteFile(filepath.Join(confDir, targetName), []byte(text), 0o640)
	}

	if err := render("app.conf", "app.conf"); err != nil {
		return err
	}
	if err := render("api.conf", "api.conf"); err != nil {
		return err
	}
	if err := render("kc.conf", "kc.conf"); err != nil {
		return err
	}

	optional := map[string]string{"grafana": "grafana.conf", "kuma": "kuma.conf"}
	for module, file := range optional {
		path := filepath.Join(confDir, file)
		if contains(modules, module) {
			if err := render(file, file); err != nil {
				return err
			}
		} else {
			_ = os.Remove(path)
		}
	}
	return nil
}
