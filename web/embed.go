package web

import "embed"

// FS embeds the index.html template file.
//
//go:embed index.html
var FS embed.FS
