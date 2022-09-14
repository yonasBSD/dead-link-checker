package reports

import "embed"

//go:embed simple_nl.html.go.tmpl
//go:embed technical_en.html.go.tmpl
var Reports embed.FS
