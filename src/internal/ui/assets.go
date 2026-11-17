package ui

import (
	"embed"
)

//go:embed templates/**/*.html
var UiFS embed.FS
