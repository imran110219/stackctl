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

func writeCompose(cfg EnvConfig, enabledModules []string) error {
	data := cfg.RenderData()

	rendered, err := renderFile("base/compose.base.yml", data)
	if err != nil {
		return err
	}

	merged := map[string]any{}
	if err := yaml.Unmarshal([]byte(rendered), &merged); err != nil {
		return err
	}

	// Only merge enabled modules, not all modules in the catalog.
	templateFS := getTemplatesFS()
	for _, module := range enabledModules {
		modPath := filepath.Join("modules", module, "compose.yml")
		// Check if module compose file exists
		if _, err := fs.Stat(templateFS, modPath); errors.Is(err, fs.ErrNotExist) {
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

func syncModuleAssets(cfg EnvConfig) error {
	templateFS := getTemplatesFS()
	entries, err := fs.ReadDir(templateFS, "modules")
	if err != nil {
		// If modules directory doesn't exist, that's okay
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		moduleName := entry.Name()
		srcDir := filepath.Join("modules", moduleName)
		dstDir := filepath.Join(cfg.EnvDir, moduleName)

		err := fs.WalkDir(templateFS, srcDir, func(path string, d fs.DirEntry, walkErr error) error {
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
				// Don't overwrite existing files
				return nil
			}

			// Read from embedded FS and write to target
			content, err := fs.ReadFile(templateFS, path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, content, 0o640)
		})
		if err != nil {
			return fmt.Errorf("sync module assets for %s: %w", moduleName, err)
		}
	}
	return nil
}

func ComposeBaseArgs(cfg EnvConfig) []string {
	return []string{
		"compose",
		"-f", filepath.Join(cfg.EnvDir, "compose.yml"),
		"-f", filepath.Join(cfg.EnvDir, "compose.override.yml"),
		"--env-file", filepath.Join(cfg.EnvDir, ".env"),
		"-p", cfg.EnvName,
	}
}

func ComposeServiceExists(cfg EnvConfig, service string) bool {
	args := ComposeBaseArgs(cfg)
	args = append(args, "config", "--services")
	out, err := RunCmdCapture("docker", args...)
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

func ComposeServiceRunning(cfg EnvConfig, service string) bool {
	args := ComposeBaseArgs(cfg)
	args = append(args, "ps", "-q", service)
	out, err := RunCmdCapture("docker", args...)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}
