package stackctl

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func runBackup(cfg EnvConfig) error {
	envMap, err := ReadDotEnv(filepath.Join(cfg.EnvDir, ".env"))
	if err != nil {
		return err
	}

	backupDir := filepath.Join(cfg.BackupRoot, cfg.EnvName)
	if err := ensureDir(backupDir, 0o750); err != nil {
		return err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")

	if err := backupIfRunning(cfg, "postgres", fmt.Sprintf("postgres_%s.sql.gz", ts),
		`PGPASSWORD="$POSTGRES_PASSWORD" pg_dumpall -U "$POSTGRES_USER"`); err != nil {
		return err
	}
	if err := backupIfRunning(cfg, "mariadb", fmt.Sprintf("mariadb_%s.sql.gz", ts),
		`mysqldump --all-databases -uroot -p"$MYSQL_ROOT_PASSWORD"`); err != nil {
		return err
	}

	resticRepo := envMap["RESTIC_REPOSITORY"]
	resticPass := envMap["RESTIC_PASSWORD"]
	if resticRepo != "" && resticPass != "" {
		fmt.Println("running optional restic push")
		cmd := exec.Command("restic", "backup", backupDir,
			filepath.Join(cfg.DataRoot, cfg.EnvName),
			filepath.Join(cfg.StackRoot, cfg.EnvName))
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

// backupIfRunning pipes the dump command output through Go's gzip writer
// instead of constructing a shell pipeline, eliminating shell interpolation.
func backupIfRunning(cfg EnvConfig, service, outName, dumpCmd string) error {
	if !ComposeServiceExists(cfg, service) {
		fmt.Printf("skip %s dump (service not defined)\n", service)
		return nil
	}
	if !ComposeServiceRunning(cfg, service) {
		fmt.Printf("skip %s dump (service not running)\n", service)
		return nil
	}

	outPath := filepath.Join(cfg.BackupRoot, cfg.EnvName, outName)

	args := append(ComposeBaseArgs(cfg), "exec", "-T", service, "sh", "-c", dumpCmd)
	cmd := exec.Command("docker", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%s dump setup failed: %w", service, err)
	}
	cmd.Stderr = os.Stderr

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create backup file: %w", err)
	}
	defer outFile.Close()

	gz := gzip.NewWriter(outFile)

	if err := cmd.Start(); err != nil {
		gz.Close()
		return fmt.Errorf("%s dump start failed: %w", service, err)
	}

	if _, err := io.Copy(gz, stdout); err != nil {
		gz.Close()
		return fmt.Errorf("%s dump copy failed: %w", service, err)
	}

	if err := gz.Close(); err != nil {
		return fmt.Errorf("%s gzip close failed: %w", service, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s dump failed: %w", service, err)
	}

	fmt.Printf("wrote %s\n", outPath)
	return nil
}
