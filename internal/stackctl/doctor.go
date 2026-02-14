package stackctl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

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
