package generator

import "embed"

var (
	//go:embed templates/*.tmpl
	templateFS embed.FS
)
