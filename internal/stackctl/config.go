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

type EnvConfig struct {
	EnvName    string
	StackRoot  string
	DataRoot   string
	BackupRoot string
	EnvDir     string
	Domain     string
	Email      string
}

func (cfg EnvConfig) RenderData() RenderData {
	return RenderData{
		Env:         cfg.EnvName,
		Domain:      cfg.Domain,
		Email:       cfg.Email,
		NetworkName: cfg.EnvName + "_net",
		StackRoot:   cfg.StackRoot,
		DataRoot:    cfg.DataRoot,
		BackupRoot:  cfg.BackupRoot,
	}
}

func LoadEnvConfig(env string) (EnvConfig, error) {
	env = strings.TrimSpace(env)
	if env != "dev" && env != "qa" && env != "prod" {
		return EnvConfig{}, errors.New("--env must be one of: dev, qa, prod")
	}
	stackRoot := GetStackRoot()
	cfg := EnvConfig{
		EnvName:    env,
		StackRoot:  stackRoot,
		DataRoot:   getDataRoot(),
		BackupRoot: getBackupRoot(),
		EnvDir:     filepath.Join(stackRoot, env),
	}
	return cfg, nil
}

func HydrateFromDotEnv(cfg *EnvConfig) error {
	m, err := ReadDotEnv(filepath.Join(cfg.EnvDir, ".env"))
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

func ReadDotEnv(path string) (map[string]string, error) {
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

func WriteDotEnv(path string, vars map[string]string) error {
	// Read original file to preserve comments and ordering
	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist, write all vars
		var b strings.Builder
		for k, v := range vars {
			b.WriteString(k + "=" + v + "\n")
		}
		return os.WriteFile(path, []byte(b.String()), 0o640)
	}
	defer file.Close()

	written := map[string]bool{}
	var lines []string
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			lines = append(lines, line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		if newVal, ok := vars[key]; ok {
			lines = append(lines, key+"="+newVal)
			written[key] = true
		} else {
			lines = append(lines, line)
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	file.Close()

	// Append any new keys that weren't in original file
	for k, v := range vars {
		if !written[k] {
			lines = append(lines, k+"="+v)
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0o640)
}

func DetectEnvironments() []string {
	stackRoot := GetStackRoot()
	envs := []string{}
	for _, name := range []string{"dev", "qa", "prod"} {
		if DirExists(filepath.Join(stackRoot, name)) {
			envs = append(envs, name)
		}
	}
	return envs
}

func GetStackRoot() string {
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
