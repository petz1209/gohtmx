package ui

import (
	"fmt"
	"html/template"
	"time"
)

var funcMap = template.FuncMap{
	"str": func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	},

	"num": func(n *int) string {
		if n == nil {
			return ""
		}
		return fmt.Sprintf("%d", *n)
	},
	"fnum": func(n *float64) string {
		if n == nil {
			return ""
		}
		return fmt.Sprintf("%.2f", *n)
	},

	"date": func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return (*t).Format("2006-01-02")
	},
}
