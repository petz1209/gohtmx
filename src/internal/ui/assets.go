package ui

import "embed"

// files holds the global assets shared by every feature:
//
//	static/    served under /static/ (app.css, app.js, htmx, logos, ...)
//	templates/ layout, partials and components, available in every feature's template set
//
// Embedding whole directories includes all subdirectories recursively
// (unlike "**" in a pattern, which only matches a single directory level).
//
//go:embed static templates
var files embed.FS
