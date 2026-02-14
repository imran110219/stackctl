package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultStackRoot  = "/srv/stack"
	defaultDataRoot   = "/srv/data"
	defaultBackupRoot = "/srv/backups"
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

type envConfig struct {
	EnvName    string
	StackRoot  string
	DataRoot   string
	BackupRoot string
	EnvDir     string
	Domain     string
	Email      string
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = runInit(args)
	case "enable":
		err = runEnableDisable(args, true)
	case "disable":
		err = runEnableDisable(args, false)
	case "status":
		err = runStatus(args)
	case "apply":
		err = runApply(args)
	case "backup":
		err = runBackup(args)
	case "doctor":
		err = runDoctor()
	case "help", "--help", "-h":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`stackctl - new VM to production-ready Docker Compose stack

Usage:
  stackctl init --env prod|devqa [--domain example.com] [--email admin@example.com]
  stackctl enable <module> --env prod|devqa
  stackctl disable <module> --env prod|devqa
  stackctl status --env prod|devqa
  stackctl apply --env prod|devqa
  stackctl backup --env prod|devqa
  stackctl doctor

Available modules:`)

	names := sortedModuleNames()
	for _, name := range names {
		m := moduleCatalog[name]
		ports := "-"
		if len(m.Ports) > 0 {
			ports = strings.Join(m.Ports, ",")
		}
		fmt.Printf("  - %-14s %-45s ports: %s\n", m.Name, m.Description, ports)
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	env := fs.String("env", "", "environment name: prod or devqa")
	domain := fs.String("domain", "example.com", "base domain")
	email := fs.String("email", "admin@example.com", "ops email")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}
	cfg.Domain = *domain
	cfg.Email = *email

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

	modules, err := loadEnabledModules(cfg)
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

func runEnableDisable(args []string, enable bool) error {
	if len(args) == 0 {
		return errors.New("module is required")
	}
	module := args[0]
	if _, ok := moduleCatalog[module]; !ok {
		return fmt.Errorf("unknown module: %s", module)
	}

	fs := flag.NewFlagSet("toggle", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	current, err := loadEnabled(cfg)
	if err != nil {
		return err
	}

	changed := false
	if enable {
		if !contains(current.Modules, module) {
			current.Modules = append(current.Modules, module)
			changed = true
		}
	} else {
		filtered := make([]string, 0, len(current.Modules))
		for _, item := range current.Modules {
			if item != module {
				filtered = append(filtered, item)
			}
		}
		if len(filtered) != len(current.Modules) {
			current.Modules = filtered
			changed = true
		}
	}

	sort.Strings(current.Modules)
	if err := writeEnabled(cfg, current); err != nil {
		return err
	}

	verb := "already disabled"
	if enable {
		verb = "already enabled"
	}
	if changed {
		if enable {
			verb = "enabled"
		} else {
			verb = "disabled"
		}
	}

	fmt.Printf("%s %s for %s\n", module, verb, cfg.EnvName)
	fmt.Printf("run: stackctl apply --env %s\n", cfg.EnvName)
	return nil
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	modules, err := loadEnabledModules(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("environment: %s\n", cfg.EnvName)
	fmt.Printf("path: %s\n", cfg.EnvDir)
	fmt.Printf("enabled modules: %s\n", strings.Join(modules, ", "))

	composeArgs := composeBaseArgs(cfg)
	composeArgs = append(composeArgs, "ps")
	output, cmdErr := runCmdCapture("docker", composeArgs...)
	if cmdErr != nil {
		fmt.Println("docker compose status unavailable:")
		fmt.Println(strings.TrimSpace(output))
		return nil
	}
	fmt.Println(output)
	return nil
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	modules, err := loadEnabledModules(cfg)
	if err != nil {
		return err
	}

	if err := hydrateFromDotEnv(&cfg); err != nil {
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
	if err := writeSystemdFiles(cfg); err != nil {
		return err
	}

	composeArgs := composeBaseArgs(cfg)
	for _, module := range modules {
		composeArgs = append(composeArgs, "--profile", module)
	}
	composeArgs = append(composeArgs, "up", "-d", "--remove-orphans")

	if err := runCmdStream("docker", composeArgs...); err != nil {
		return err
	}

	fmt.Printf("applied %s with modules: %s\n", cfg.EnvName, strings.Join(modules, ", "))
	return nil
}

func runBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	envMap, err := readDotEnv(filepath.Join(cfg.EnvDir, ".env"))
	if err != nil {
		return err
	}

	if err := ensureDir(filepath.Join(cfg.BackupRoot, cfg.EnvName), 0o750); err != nil {
		return err
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	backupDir := filepath.Join(cfg.BackupRoot, cfg.EnvName)

	if err := backupIfRunning(cfg, "postgres", fmt.Sprintf("postgres_%s.sql.gz", ts),
		"PGPASSWORD=\"$POSTGRES_PASSWORD\" pg_dumpall -U \"$POSTGRES_USER\""); err != nil {
		return err
	}
	if err := backupIfRunning(cfg, "mariadb", fmt.Sprintf("mariadb_%s.sql.gz", ts),
		"mysqldump --all-databases -uroot -p\"$MYSQL_ROOT_PASSWORD\""); err != nil {
		return err
	}

	resticRepo := envMap["RESTIC_REPOSITORY"]
	resticPass := envMap["RESTIC_PASSWORD"]
	if resticRepo != "" && resticPass != "" {
		fmt.Println("running optional restic push")
		cmd := exec.Command("restic", "backup", backupDir, filepath.Join(cfg.DataRoot, cfg.EnvName), filepath.Join(cfg.StackRoot, cfg.EnvName))
		cmd.Env = append(os.Environ(),
			"RESTIC_REPOSITORY="+resticRepo,
			"RESTIC_PASSWORD="+resticPass,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("restic backup failed: %w", err)
		}
	} else {
		fmt.Println("restic skipped (RESTIC_REPOSITORY/RESTIC_PASSWORD not set)")
	}

	return nil
}

func runDoctor() error {
	fmt.Println("stackctl doctor")
	fmt.Printf("runtime: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	checks := []struct {
		name string
		fn   func() error
	}{
		{"docker binary", func() error {
			_, err := exec.LookPath("docker")
			return err
		}},
		{"docker compose", func() error {
			_, err := runCmdCapture("docker", "compose", "version")
			return err
		}},
		{"docker daemon", func() error {
			_, err := runCmdCapture("docker", "info")
			return err
		}},
		{"/srv/stack writable", func() error {
			return writableCheck(getStackRoot())
		}},
		{"/srv/data writable", func() error {
			return writableCheck(getDataRoot())
		}},
		{"disk space >= 5GiB on /srv", func() error {
			return diskCheck("/srv", 5)
		}},
		{"ports 80/443 status", func() error {
			out, err := runCmdCapture("ss", "-ltn")
			if err != nil {
				return err
			}
			if strings.Contains(out, ":80 ") || strings.Contains(out, ":443 ") {
				return fmt.Errorf("ports 80/443 already in use")
			}
			return nil
		}},
	}

	for _, check := range checks {
		if err := check.fn(); err != nil {
			fmt.Printf("[WARN] %s: %v\n", check.name, err)
		} else {
			fmt.Printf("[ OK ] %s\n", check.name)
		}
	}
	return nil
}

func loadEnvConfig(env string) (envConfig, error) {
	env = strings.TrimSpace(env)
	if env != "prod" && env != "devqa" {
		return envConfig{}, errors.New("--env must be prod or devqa")
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

func ensureEnvDirs(cfg envConfig) error {
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

func ensureDefaultEnabled(cfg envConfig) error {
	path := filepath.Join(cfg.EnvDir, "enabled.yml")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	def := enabledConfig{Modules: []string{}}
	return writeEnabled(cfg, def)
}

func ensureDotEnv(cfg envConfig) error {
	target := filepath.Join(cfg.EnvDir, ".env")
	if _, err := os.Stat(target); err == nil {
		return nil
	}

	tplPath := filepath.Join(findTemplatesDir(), ".env.example")
	content, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read .env template: %w", err)
	}

	text := string(content)
	text = strings.ReplaceAll(text, "{{ENV}}", cfg.EnvName)
	text = strings.ReplaceAll(text, "{{DOMAIN}}", cfg.Domain)
	text = strings.ReplaceAll(text, "{{EMAIL}}", cfg.Email)
	return os.WriteFile(target, []byte(text), 0o640)
}

func ensureComposeOverride(cfg envConfig) error {
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

func writeCompose(cfg envConfig, enabledModules []string) error {
	templates := findTemplatesDir()
	basePath := filepath.Join(templates, "base", "compose.base.yml")
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return err
	}

	rendered := strings.ReplaceAll(string(baseData), "{{ENV}}", cfg.EnvName)
	rendered = strings.ReplaceAll(rendered, "{{NETWORK_NAME}}", cfg.EnvName+"_net")

	merged := map[string]any{}
	if err := yaml.Unmarshal([]byte(rendered), &merged); err != nil {
		return err
	}

	for _, module := range sortedModuleNames() {
		modPath := filepath.Join(templates, "modules", module, "compose.yml")
		if _, err := os.Stat(modPath); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		modData, err := os.ReadFile(modPath)
		if err != nil {
			return err
		}
		modRendered := strings.ReplaceAll(string(modData), "{{ENV}}", cfg.EnvName)
		modRendered = strings.ReplaceAll(modRendered, "{{NETWORK_NAME}}", cfg.EnvName+"_net")
		var overlay map[string]any
		if err := yaml.Unmarshal([]byte(modRendered), &overlay); err != nil {
			return fmt.Errorf("parse module %s compose: %w", module, err)
		}
		deepMerge(merged, overlay)
	}

	if _, ok := merged["x-stackctl"]; !ok {
		merged["x-stackctl"] = map[string]any{}
	}
	x := merged["x-stackctl"].(map[string]any)
	x["enabled_modules"] = enabledModules
	x["generated_at"] = time.Now().UTC().Format(time.RFC3339)

	out, err := yaml.Marshal(merged)
	if err != nil {
		return err
	}

	target := filepath.Join(cfg.EnvDir, "compose.yml")
	return os.WriteFile(target, out, 0o640)
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

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		existing, exists := dst[k]
		if !exists {
			dst[k] = v
			continue
		}

		dstMap, dstMapOK := existing.(map[string]any)
		srcMap, srcMapOK := v.(map[string]any)
		if dstMapOK && srcMapOK {
			deepMerge(dstMap, srcMap)
			dst[k] = dstMap
			continue
		}
		dst[k] = v
	}
}

func writeNginxConfs(cfg envConfig, modules []string) error {
	if cfg.Domain == "" {
		if err := hydrateFromDotEnv(&cfg); err != nil {
			return err
		}
	}
	confDir := filepath.Join(cfg.EnvDir, "nginx", "conf.d")
	if err := ensureDir(confDir, 0o750); err != nil {
		return err
	}

	templates := findTemplatesDir()
	render := func(templateName, targetName string) error {
		inPath := filepath.Join(templates, "nginx", templateName)
		content, err := os.ReadFile(inPath)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(content), "{{DOMAIN}}", cfg.Domain)
		text = strings.ReplaceAll(text, "{{ENV}}", cfg.EnvName)
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

func writeBackupScript(cfg envConfig) error {
	templates := findTemplatesDir()
	tpl := filepath.Join(templates, "systemd", "backup-now.sh")
	b, err := os.ReadFile(tpl)
	if err != nil {
		return err
	}
	text := string(b)
	text = strings.ReplaceAll(text, "{{ENV}}", cfg.EnvName)
	text = strings.ReplaceAll(text, "{{STACK_ROOT}}", cfg.StackRoot)
	text = strings.ReplaceAll(text, "{{DATA_ROOT}}", cfg.DataRoot)
	text = strings.ReplaceAll(text, "{{BACKUP_ROOT}}", cfg.BackupRoot)

	target := filepath.Join(cfg.EnvDir, "backup-now.sh")
	if err := os.WriteFile(target, []byte(text), 0o750); err != nil {
		return err
	}
	return nil
}

func syncModuleAssets(cfg envConfig) error {
	templates := findTemplatesDir()
	modulesDir := filepath.Join(templates, "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		moduleName := entry.Name()
		srcDir := filepath.Join(modulesDir, moduleName)
		dstDir := filepath.Join(cfg.EnvDir, moduleName)

		err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(srcDir, path)
			if err != nil {
				return err
			}
			if rel == "." {
				return nil
			}
			if d.IsDir() {
				return ensureDir(filepath.Join(dstDir, rel), 0o750)
			}
			if filepath.Base(path) == "compose.yml" {
				return nil
			}

			target := filepath.Join(dstDir, rel)
			if _, err := os.Stat(target); err == nil {
				return nil
			}
			return copyFile(path, target)
		})
		if err != nil {
			return fmt.Errorf("sync module assets for %s: %w", moduleName, err)
		}
	}
	return nil
}

func writeSystemdFiles(cfg envConfig) error {
	templates := findTemplatesDir()
	targetDir := filepath.Join(cfg.EnvDir, "systemd")
	if err := ensureDir(targetDir, 0o750); err != nil {
		return err
	}

	type filePair struct {
		in  string
		out string
	}
	files := []filePair{
		{in: "stackctl-env.service", out: fmt.Sprintf("stackctl-%s.service", cfg.EnvName)},
		{in: "stackctl-backup.service", out: fmt.Sprintf("stackctl-backup-%s.service", cfg.EnvName)},
		{in: "stackctl-backup.timer", out: fmt.Sprintf("stackctl-backup-%s.timer", cfg.EnvName)},
	}

	for _, pair := range files {
		inPath := filepath.Join(templates, "systemd", pair.in)
		content, err := os.ReadFile(inPath)
		if err != nil {
			return err
		}
		text := string(content)
		text = strings.ReplaceAll(text, "{{ENV}}", cfg.EnvName)
		text = strings.ReplaceAll(text, "{{STACK_ROOT}}", cfg.StackRoot)
		target := filepath.Join(targetDir, pair.out)
		if err := os.WriteFile(target, []byte(text), 0o644); err != nil {
			return err
		}
	}

	if os.Geteuid() == 0 {
		for _, pair := range files {
			src := filepath.Join(targetDir, pair.out)
			dst := filepath.Join("/etc/systemd/system", pair.out)
			b, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dst, b, 0o644); err != nil {
				return err
			}
		}
		_ = runCmdStream("systemctl", "daemon-reload")
		_ = runCmdStream("systemctl", "enable", fmt.Sprintf("stackctl-%s.service", cfg.EnvName))
		_ = runCmdStream("systemctl", "enable", fmt.Sprintf("stackctl-backup-%s.timer", cfg.EnvName))
	}
	return nil
}

func backupIfRunning(cfg envConfig, service, outName, shellCmd string) error {
	if !composeServiceExists(cfg, service) {
		fmt.Printf("skip %s dump (service not defined)\n", service)
		return nil
	}
	if !composeServiceRunning(cfg, service) {
		fmt.Printf("skip %s dump (service not running)\n", service)
		return nil
	}

	outPath := filepath.Join(cfg.BackupRoot, cfg.EnvName, outName)
	compose := strings.Join(composeBaseArgs(cfg), " ")
	cmdline := fmt.Sprintf("docker %s exec -T %s sh -c '%s' | gzip -c > %s", compose, service, shellCmd, shellEscape(outPath))
	if err := runCmdStream("sh", "-c", cmdline); err != nil {
		return fmt.Errorf("%s dump failed: %w", service, err)
	}
	fmt.Printf("wrote %s\n", outPath)
	return nil
}

func composeServiceExists(cfg envConfig, service string) bool {
	args := composeBaseArgs(cfg)
	args = append(args, "config", "--services")
	out, err := runCmdCapture("docker", args...)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == service {
			return true
		}
	}
	return false
}

func composeServiceRunning(cfg envConfig, service string) bool {
	args := composeBaseArgs(cfg)
	args = append(args, "ps", "-q", service)
	out, err := runCmdCapture("docker", args...)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

func composeBaseArgs(cfg envConfig) []string {
	return []string{
		"compose",
		"-f", filepath.Join(cfg.EnvDir, "compose.yml"),
		"-f", filepath.Join(cfg.EnvDir, "compose.override.yml"),
		"--env-file", filepath.Join(cfg.EnvDir, ".env"),
		"-p", cfg.EnvName,
	}
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

func writableCheck(dir string) error {
	if err := ensureDir(dir, 0o750); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "stackctl-write-check-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}

func diskCheck(path string, minGiB uint64) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}
	free := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	if free < minGiB {
		return fmt.Errorf("free space %dGiB < %dGiB", free, minGiB)
	}
	return nil
}

func ensureDir(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func runCmdCapture(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runCmdStream(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}

func sortedModuleNames() []string {
	names := make([]string, 0, len(moduleCatalog))
	for name := range moduleCatalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func findTemplatesDir() string {
	if custom := strings.TrimSpace(os.Getenv("STACKCTL_TEMPLATES")); custom != "" {
		return custom
	}

	exe, err := os.Executable()
	if err == nil {
		binDir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(binDir, "..", "templates"),
			filepath.Join(binDir, "templates"),
		}
		for _, c := range candidates {
			if dirExists(c) {
				return c
			}
		}
	}

	cwd, err := os.Getwd()
	if err == nil {
		c := filepath.Join(cwd, "templates")
		if dirExists(c) {
			return c
		}
	}

	home, _ := os.UserHomeDir()
	fallbacks := []string{
		"/usr/local/share/stackctl/templates",
		filepath.Join(home, ".stackctl", "repo", "templates"),
	}
	for _, c := range fallbacks {
		if dirExists(c) {
			return c
		}
	}
	return "templates"
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func shellEscape(in string) string {
	return "'" + strings.ReplaceAll(in, "'", "'\\''") + "'"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := ensureDir(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
