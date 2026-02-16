package stackctl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

type CheckResult struct {
	Name string
	OK   bool
	Err  error
}

func RunChecks() []CheckResult {
	checks := []struct {
		name string
		fn   func() error
	}{
		{"docker binary", func() error {
			_, err := exec.LookPath("docker")
			return err
		}},
		{"docker compose", func() error {
			_, err := RunCmdCapture("docker", "compose", "version")
			return err
		}},
		{"docker daemon", func() error {
			_, err := RunCmdCapture("docker", "info")
			return err
		}},
		{"/srv/stack writable", func() error {
			return writableCheck(GetStackRoot())
		}},
		{"/srv/data writable", func() error {
			return writableCheck(getDataRoot())
		}},
		{"disk space >= 5GiB on /srv", func() error {
			return diskCheck("/srv", 5)
		}},
		{"ports 80/443 status", func() error {
			out, err := RunCmdCapture("ss", "-ltn")
			if err != nil {
				return err
			}
			if strings.Contains(out, ":80 ") || strings.Contains(out, ":443 ") {
				return fmt.Errorf("ports 80/443 already in use")
			}
			return nil
		}},
	}

	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		err := check.fn()
		results = append(results, CheckResult{
			Name: check.name,
			OK:   err == nil,
			Err:  err,
		})
	}
	return results
}

func RunDoctor() error {
	fmt.Println("stackctl doctor")
	fmt.Printf("runtime: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	results := RunChecks()
	for _, r := range results {
		if r.OK {
			fmt.Printf("[ OK ] %s\n", r.Name)
		} else {
			fmt.Printf("[WARN] %s: %v\n", r.Name, r.Err)
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
