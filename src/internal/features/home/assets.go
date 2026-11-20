package home

import "embed"

//go:embed templates/**/*.html
var htmlFiles embed.FS

//go:embed templates/**/*js
var jsFiles embed.FS

//go:embed templates/**/*.css
var cssFiles embed.FS
