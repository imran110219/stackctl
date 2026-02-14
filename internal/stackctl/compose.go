package stackctl

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func writeCompose(cfg envConfig, enabledModules []string) error {
	templates := findTemplatesDir()
	data := cfg.renderData()

	basePath := filepath.Join(templates, "base", "compose.base.yml")
	rendered, err := renderFile(basePath, data)
	if err != nil {
		return err
	}

	merged := map[string]any{}
	if err := yaml.Unmarshal([]byte(rendered), &merged); err != nil {
		return err
	}

	// Only merge enabled modules, not all modules in the catalog.
	for _, module := range enabledModules {
		modPath := filepath.Join(templates, "modules", module, "compose.yml")
		if _, err := os.Stat(modPath); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		modRendered, err := renderFile(modPath, data)
		if err != nil {
			return fmt.Errorf("render module %s compose: %w", module, err)
		}
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
			continue
		}

		dstSlice, dstSliceOK := existing.([]any)
		srcSlice, srcSliceOK := v.([]any)
		if dstSliceOK && srcSliceOK {
			dst[k] = append(dstSlice, srcSlice...)
			continue
		}

		dst[k] = v
	}
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

func composeBaseArgs(cfg envConfig) []string {
	return []string{
		"compose",
		"-f", filepath.Join(cfg.EnvDir, "compose.yml"),
		"-f", filepath.Join(cfg.EnvDir, "compose.override.yml"),
		"--env-file", filepath.Join(cfg.EnvDir, ".env"),
		"-p", cfg.EnvName,
	}
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
