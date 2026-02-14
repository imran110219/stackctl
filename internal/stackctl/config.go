package stackctl

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultStackRoot  = "/srv/stack"
	defaultDataRoot   = "/srv/data"
	defaultBackupRoot = "/srv/backups"
)

type envConfig struct {
	EnvName    string
	StackRoot  string
	DataRoot   string
	BackupRoot string
	EnvDir     string
	Domain     string
	Email      string
}

func (cfg envConfig) renderData() renderData {
	return renderData{
		Env:         cfg.EnvName,
		Domain:      cfg.Domain,
		Email:       cfg.Email,
		NetworkName: cfg.EnvName + "_net",
		StackRoot:   cfg.StackRoot,
		DataRoot:    cfg.DataRoot,
		BackupRoot:  cfg.BackupRoot,
	}
}

func loadEnvConfig(env string) (envConfig, error) {
	env = strings.TrimSpace(env)
	if env != "dev" && env != "qa" && env != "prod" {
		return envConfig{}, errors.New("--env must be one of: dev, qa, prod")
	}
	stackRoot := getStackRoot()
	cfg := envConfig{
		EnvName:    env,
		StackRoot:  stackRoot,
		DataRoot:   getDataRoot(),
		BackupRoot: getBackupRoot(),
		EnvDir:     filepath.Join(stackRoot, env),
	}
	return cfg, nil
}

func hydrateFromDotEnv(cfg *envConfig) error {
	m, err := readDotEnv(filepath.Join(cfg.EnvDir, ".env"))
	if err != nil {
		return err
	}
	if cfg.Domain == "" {
		cfg.Domain = m["DOMAIN"]
	}
	if cfg.Email == "" {
		cfg.Email = m["ADMIN_EMAIL"]
	}
	return nil
}

func readDotEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := map[string]string{}
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		vars[k] = v
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return vars, nil
}

func getStackRoot() string {
	if v := strings.TrimSpace(os.Getenv("STACKCTL_STACK_ROOT")); v != "" {
		return v
	}
	return defaultStackRoot
}

func getDataRoot() string {
	if v := strings.TrimSpace(os.Getenv("STACKCTL_DATA_ROOT")); v != "" {
		return v
	}
	return defaultDataRoot
}

func getBackupRoot() string {
	if v := strings.TrimSpace(os.Getenv("STACKCTL_BACKUP_ROOT")); v != "" {
		return v
	}
	return defaultBackupRoot
}
