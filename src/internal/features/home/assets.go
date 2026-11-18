package home

import "embed"

//go:embed templates/**/*.html
var htmlFiles embed.FS

//go:embed templates/scripts/*js
var jsFiles embed.FS
