package stackctl

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeSystemdFiles(cfg EnvConfig) error {
	templates := findTemplatesDir()
	data := cfg.RenderData()
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
		text, err := renderFile(inPath, data)
		if err != nil {
			return fmt.Errorf("render systemd %s: %w", pair.in, err)
		}
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

func writeBackupScript(cfg EnvConfig) error {
	templates := findTemplatesDir()
	data := cfg.RenderData()
	tplPath := filepath.Join(templates, "systemd", "backup-now.sh")
	text, err := renderFile(tplPath, data)
	if err != nil {
		return fmt.Errorf("render backup script: %w", err)
	}
	target := filepath.Join(cfg.EnvDir, "backup-now.sh")
	return os.WriteFile(target, []byte(text), 0o750)
}
