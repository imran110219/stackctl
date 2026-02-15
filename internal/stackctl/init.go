package stackctl

import (
	"fmt"
	"os"
	"path/filepath"
)

func RunInit(cfg EnvConfig) error {
	if err := ensureDir(cfg.EnvDir, 0o750); err != nil {
		return err
	}
	if err := ensureDir(cfg.DataRoot, 0o750); err != nil {
		return err
	}
	if err := ensureDir(cfg.BackupRoot, 0o750); err != nil {
		return err
	}
	if err := ensureDir(filepath.Join(cfg.BackupRoot, cfg.EnvName), 0o750); err != nil {
		return err
	}

	if err := ensureEnvDirs(cfg); err != nil {
		return err
	}
	if err := ensureDefaultEnabled(cfg); err != nil {
		return err
	}
	if err := ensureDotEnv(cfg); err != nil {
		return err
	}
	if err := ensureComposeOverride(cfg); err != nil {
		return err
	}

	modules, err := LoadEnabledModules(cfg)
	if err != nil {
		return err
	}

	if err := writeCompose(cfg, modules); err != nil {
		return err
	}
	if err := syncModuleAssets(cfg); err != nil {
		return err
	}
	if err := writeNginxConfs(cfg, modules); err != nil {
		return err
	}
	if err := writeBackupScript(cfg); err != nil {
		return err
	}
	if err := writeSystemdFiles(cfg); err != nil {
		return err
	}

	fmt.Printf("initialized %s at %s\n", cfg.EnvName, cfg.EnvDir)
	fmt.Printf("next: stackctl apply --env %s\n", cfg.EnvName)
	return nil
}

func ensureEnvDirs(cfg EnvConfig) error {
	dirs := []string{
		cfg.EnvDir,
		filepath.Join(cfg.EnvDir, "nginx", "conf.d"),
		filepath.Join(cfg.EnvDir, "systemd"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "nginx"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "frontend"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "backend"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "keycloak"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "postgres"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "mariadb"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "prometheus"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "grafana"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "loki"),
		filepath.Join(cfg.DataRoot, cfg.EnvName, "kuma"),
		filepath.Join(cfg.BackupRoot, cfg.EnvName),
	}
	for _, dir := range dirs {
		if err := ensureDir(dir, 0o750); err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultEnabled(cfg EnvConfig) error {
	path := filepath.Join(cfg.EnvDir, "enabled.yml")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	def := EnabledConfig{Modules: []string{}}
	return WriteEnabled(cfg, def)
}

func ensureDotEnv(cfg EnvConfig) error {
	target := filepath.Join(cfg.EnvDir, ".env")
	if _, err := os.Stat(target); err == nil {
		return nil
	}

	tplPath := filepath.Join(findTemplatesDir(), ".env.example")
	data := cfg.RenderData()
	text, err := renderFile(tplPath, data)
	if err != nil {
		return fmt.Errorf("render .env template: %w", err)
	}
	return os.WriteFile(target, []byte(text), 0o640)
}

func ensureComposeOverride(cfg EnvConfig) error {
	target := filepath.Join(cfg.EnvDir, "compose.override.yml")
	if _, err := os.Stat(target); err == nil {
		return nil
	}

	tplPath := filepath.Join(findTemplatesDir(), "base", "compose.override.yml")
	content, err := os.ReadFile(tplPath)
	if err != nil {
		return err
	}
	return os.WriteFile(target, content, 0o640)
}
