# Markdown Browser (`md-browser`)

A high-performance, responsive, and beautiful local Markdown and source code browser built in Go and HTML/JS. It lets you instantly navigate, read, and render local Markdown files and code repositories inside a modern dual-pane web application with **GitHub Markdown Light** styling.

---

## Key Features

- **GitHub Markdown Theme**: Markdown files are rendered perfectly using the official GitHub light style inside a clean paper sheet, while the main application frame keeps a gorgeous, eye-friendly dark theme.
- **Pristine GFM Syntax**: Uses the highly compliant standard `goldmark` engine with GFM extensions (strikethroughs, robust tables, syntax code-blocks).
- **Standalone Source Code & Syntax Highlighting**: Seamlessly views standard source and configuration files (e.g., `.go`, `.json`, `.yaml`, `.mod`, `.sum`) with syntax highlighting powered by `Highlight.js` alongside standard line-number gutters.
- **Null-Byte Binary Sniffing**: Safely detects and identifies binary files (using null-byte inspection and mimetype sniffing) to show a beautiful file detail card instead of corrupting the layout with unreadable binary text.
- **Flicker-Free DOM Reconciliation**: Updates the directory filetree in real-time using an intelligent DOM patching algorithm. Files and folders are loaded dynamically, preserving active selections and fold states without flashing or scrolling resets.
- **Security-First Path Traversal Protection**: Employs robust verification checks on all path parameters using Go's absolute path resolution to block directory-traversal exploits.
- **Platform-Agnostic Autostart**: Automatically triggers the default browser to launch the client on start (fully configurable via flags).

---

## Project Structure

This project follows clean, production-ready Go project design practices:

```
├── cmd/
│   └── md-browser/
│       └── main.go                  # Entrypoint of the application
├── internal/
│   ├── config/
│   │   └── config.go                # Setting structure and CLI flag parsing
│   ├── markdown/
│   │   ├── parser.go                # Goldmark conversion engine wrapper
│   │   └── parser_test.go           # Markdown parser tests
│   └── server/
│       ├── handlers.go              # Business routes, path-traversal check, and sniffing
│       ├── handlers_test.go         # Traversal prevention security tests
│       └── server.go                # HTTP daemon coordinator
├── web/
│   ├── embed.go                     # Standard FS embedded assets
│   └── index.html                   # Dual-pane single-page application UI
├── install.sh                       # Dynamic local build/install script
└── md-browser.service               # Systemd user service template
```

---

## Installation

You can install `md-browser` automatically using the POSIX-compliant installation script:

```bash
chmod +x install.sh
./install.sh
```

### Dynamic Install Targets
- If `$GO_PATH` or `$GOPATH` exists, the script compiles and installs the binary to `$GOPATH/bin`.
- If no Go paths are detected, it falls back to installing under `/usr/local/bin` (and prompts for `sudo` elevation if write permissions are required).

---

## Usage

By default, executing `md-browser` launches the application as a **daemon running in the background**, so your terminal is immediately freed.

### Start the Browser
```bash
# Starts the browser in the background (default)
md-browser

# Run in the foreground instead of daemonizing
md-browser -foreground
```

### Stop the Browser
```bash
# Stops ALL currently running background instances collectively
md-browser -stop

# Stops ONLY the background instance on a custom port (e.g. 8085)
md-browser -stop -port 8085
```

### Check Running Status
To check all running instances on your system:
```bash
md-browser -status
```
Example Output:
```
Running md-browser instances:
  • PID: 683496 | Port: 8080 | Directory: /home/bas/documents
  • PID: 683512 | Port: 8081 | Directory: /home/bas/projects
```

### View Server Logs
To print the startup and runtime output logs for a background instance:
```bash
# View logs for default port 8080
md-browser -show-logs

# View logs for a custom port
md-browser -show-logs -port 8081
```

### Port & Directory Conflict Resolution

- **Same-Directory Blocker**: To avoid wasting system resources, `md-browser` automatically detects if there is already a running background instance serving the exact same local folder. If found, the app prints a notification with the active process details and exits immediately, instead of starting a redundant second server.
- **Port Scouting**: If you start `md-browser` on a different local folder and the configured port (default `8080`) is already in use, `md-browser` will automatically detect the port conflict, scout for the next available free port (e.g. `8081`), notify you on the terminal, and bind to it seamlessly.

### Command Line Flags

Customize the browser using standard command line options:

| Flag | Default | Description |
|---|---|---|
| `-dir` / `-directory` | `.` | Root directory path on disk to browse and serve. |
| `-port` | `8080` | Port to run the HTTP server on. |
| `-refresh-interval` | `5` | Active background rate (in seconds) to query filetree updates. |
| `-disable-open` | `false` | Turn off automatic web browser launch on system startup. |
| `-foreground` | `false` | Run the web server in the foreground instead of background daemonizing. |
| `-stop` | `false` | Stop the currently running background instance of the Markdown Browser. |
| `-show-logs` | `false` | Show logs of the background Markdown Browser instance on the specified port. |
| `-status` | `false` | Show currently running Markdown Browser instances (including PID, port, and root directory). |

#### Examples:
```bash
# Browse home folder documents with an 8-second refresh rate
md-browser -dir ~/Documents -refresh-interval 8

# Host on port 9090 and disable auto-browser startup
md-browser -port 9090 -disable-open
```

---

## Running as a Systemd User Service

To let the markdown browser run silently in the background, you can set it up as a standard Systemd user service:

1. **Create the configuration directory** if it doesn't exist:
   ```bash
   mkdir -p ~/.config/systemd/user/
   ```

2. **Copy the service file template**:
   ```bash
   cp md-browser.service ~/.config/systemd/user/md-browser.service
   ```

3. **Reload the systemd user daemon & start**:
   ```bash
   systemctl --user daemon-reload
   systemctl --user enable md-browser.service --now
   ```

4. **Check active status**:
   ```bash
   systemctl --user status md-browser.service
   ```

---

## Development & Test Commands

Run tests and standard tasks from the repository root:

- **Run unit test suite**:
  ```bash
  go test -v ./...
  ```
- **Format code**:
  ```bash
  go fmt ./...
  ```
- **Compile locally**:
  ```bash
  go build -o md-browser ./cmd/md-browser
  ```
- **Run without compilation**:
  ```bash
  go run ./cmd/md-browser
  ```
