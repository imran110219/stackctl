// This file embeds templates for distribution in the stackctl binary

package stackctl

import "embed"

//go:embed all:templates
var EmbeddedTemplates embed.FS
