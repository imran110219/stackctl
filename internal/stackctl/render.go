package stackctl

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"strings"
	"text/template"
)

// embeddedTemplates is initialized by InitTemplates from the main package
var embeddedTemplates embed.FS

// InitTemplates initializes the embedded templates from the main package
func InitTemplates(fsys embed.FS) {
	embeddedTemplates = fsys
}

type RenderData struct {
	Env         string
	Domain      string
	Email       string
	NetworkName string
	StackRoot   string
	DataRoot    string
	BackupRoot  string
}

// getTemplatesFS returns the filesystem to use for templates.
// If STACKCTL_TEMPLATES is set, it returns a DirFS of that path (for development).
// Otherwise, it returns the embedded templates FS.
func getTemplatesFS() fs.FS {
	if custom := strings.TrimSpace(os.Getenv("STACKCTL_TEMPLATES")); custom != "" {
		return os.DirFS(custom)
	}
	// embeddedTemplates contains "templates/" prefix from embed path
	templatesFS, err := fs.Sub(embeddedTemplates, "templates")
	if err != nil {
		// Should never happen since we embed templates/
		panic("failed to access embedded templates: " + err.Error())
	}
	return templatesFS
}

// renderFile renders a template file from the templates FS.
// The path should be relative to the templates directory (e.g., "base/compose.base.yml").
func renderFile(path string, data RenderData) (string, error) {
	templateFS := getTemplatesFS()
	content, err := fs.ReadFile(templateFS, path)
	if err != nil {
		return "", err
	}
	return renderString(string(content), data)
}

// readTemplateFile reads a file from the templates FS without rendering.
func readTemplateFile(path string) ([]byte, error) {
	templateFS := getTemplatesFS()
	return fs.ReadFile(templateFS, path)
}

func renderString(content string, data RenderData) (string, error) {
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

// findTemplatesDir is deprecated. Use getTemplatesFS() instead.
// This function is kept for backward compatibility only.
func findTemplatesDir() string {
	if custom := strings.TrimSpace(os.Getenv("STACKCTL_TEMPLATES")); custom != "" {
		return custom
	}
	// When using embedded FS, return empty string
	return ""
}
