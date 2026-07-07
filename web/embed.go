package web

import "embed"

// FS embeds the index.html, style.css, and app.js files.
//
//go:embed index.html style.css app.js
var FS embed.FS
