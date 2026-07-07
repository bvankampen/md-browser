# Description

Webbased markdown browser, which allows you to quickly read and browse markdown on local disk.

# Application

This is a golang application which starts up a webserver on localhost. This webserver show the directory structure on the left and the parsed markdown file on the right. The same markdown theme should be used as github uses.

# Critical Constraints (Do Not Miss)

- **Security (Path Traversal)**: Since the application reads and serves local files, you MUST strictly sanitize all path parameters to prevent directory traversal attacks (e.g., by resolving paths to absolute paths and checking they start with the configured root directory prefix).
- **GitHub Markdown Styling**: To match GitHub's exact styling, link the GitHub Markdown Light CSS (e.g., `https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.8.1/github-markdown-light.min.css`) and wrap the parsed HTML content inside a container with class `markdown-body` on a clean, light-themed sheet layout, while keeping the rest of the application interface dark.
- **Go Markdown Parsing**: Use `github.com/yuin/goldmark` with GFM extensions (`extension.GFM`) to perform standard-compliant markdown rendering in Go, rather than legacy packages like `blackfriday`.
- **Read-Only**: The application is strictly a read-only viewer. Do not implement any editing, writing, or deleting functionality on the filesystem.
- **Serving Local Assets**: Serve local assets (such as images, diagrams, or assets linked inside Markdown documents) securely. Translate relative references relative to the active document's parent directory and route requests through the safe-path traversal verification check before delivery.
- **Empty Directories UI**: When "Show all files" is disabled, any subdirectory containing no GFM Markdown files must display `(no md)` inline next to the directory name in italicized, muted text, and standard nested empty labels within them must be hidden.
- **Git Operations**: Only commit, amend, push, or create PRs when explicitly requested by the user.

# Architecture & Tech Stack

- **Backend**: Go standard library (`net/http`, `html/template`, `embed`). Keep it clean and simple.
- **Frontend**: Responsive split-pane layout using standard CSS Grid/Flexbox and vanilla JavaScript. Avoid complex client-side bundlers.
- **Navigation**: Dynamically traverse and display the folder structure on the left pane and display the selected rendered markdown on the right pane.
- **Directory Configuration**: Default to the current working directory as the root, but support setting a different directory path via a CLI flag (e.g., `--dir` or `--directory`).
- **Browser Autostart**: Automatically open the default web browser on launch to the server URL. Provide a CLI flag (`--disable-open` with default `false`) to turn this off if requested.
- **File Filtering**: Only show folders and markdown files (`.md`, `.markdown`) by default. Provide a configuration/toggle option to show all files.

# Developer Commands

- Initialize Go module: `go mod init github.com/bvankampen/md-browser`
- Tidy dependencies: `go mod tidy`
- Run application: `go run ./cmd/md-browser`
- Run tests: `go test ./...`
- Format code: `go fmt ./...`
