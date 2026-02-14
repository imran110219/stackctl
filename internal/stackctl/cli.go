package stackctl

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func Run(args []string) error {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "init":
		return cmdInit(cmdArgs)
	case "enable":
		return cmdEnableDisable(cmdArgs, true)
	case "disable":
		return cmdEnableDisable(cmdArgs, false)
	case "status":
		return cmdStatus(cmdArgs)
	case "apply":
		return cmdApply(cmdArgs)
	case "backup":
		return cmdBackup(cmdArgs)
	case "doctor":
		return runDoctor()
	case "help", "--help", "-h":
		usage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func usage() {
	fmt.Println(`stackctl - new VM to production-ready Docker Compose stack

Usage:
  stackctl init --env dev|qa|prod [--domain example.com] [--email admin@example.com]
  stackctl enable <module> --env dev|qa|prod
  stackctl disable <module> --env dev|qa|prod
  stackctl status --env dev|qa|prod
  stackctl apply --env dev|qa|prod
  stackctl backup --env dev|qa|prod
  stackctl doctor

Available modules:`)

	names := sortedModuleNames()
	for _, name := range names {
		m := moduleCatalog[name]
		fmt.Printf("  - %-14s %-45s ports: %s\n", m.Name, m.Description, sortedModulePorts(name))
	}
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	env := fs.String("env", "", "environment name: dev, qa, or prod")
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

	return runInit(cfg)
}

func cmdEnableDisable(args []string, enable bool) error {
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

func cmdStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	if err := hydrateFromDotEnv(&cfg); err != nil {
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

func cmdApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	if err := hydrateFromDotEnv(&cfg); err != nil {
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

func cmdBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	env := fs.String("env", "", "environment name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadEnvConfig(*env)
	if err != nil {
		return err
	}

	if err := hydrateFromDotEnv(&cfg); err != nil {
		return err
	}

	return runBackup(cfg)
}
