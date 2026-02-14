package stackctl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type renderData struct {
	Env         string
	Domain      string
	Email       string
	NetworkName string
	StackRoot   string
	DataRoot    string
	BackupRoot  string
}

func renderFile(path string, data renderData) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return renderString(string(content), data)
}

func renderString(content string, data renderData) (string, error) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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
