# aegis-cli
<img width="1103" height="245" alt="Screenshot 2026-02-24 at 5 21 27 PM" src="https://github.com/user-attachments/assets/e059598a-ca90-4870-9946-48fc8b32e7c4" />

A lightweight CLI tool for tracking deep work sessions and distraction impulses. Built in Go, runs locally, no external services.

## Install

### Quick install (macOS / Linux)

```bash
curl -sL https://raw.githubusercontent.com/nmashchenko/aegis-cli/master/install.sh | sh
```

Auto-detects your OS and architecture, downloads the latest release, and installs to `/usr/local/bin`.

### With Go

```bash
go install github.com/nmashchenko/aegis-cli@latest
```

### From source

```bash
git clone https://github.com/nmashchenko/aegis-cli.git
cd aegis-cli
go build -o aegis .
sudo mv aegis /usr/local/bin/
```

## Usage

### Start a focus task

```bash
# Live TUI mode (default) — shows timer, urge counter, keybindings
aegis start "coding"

# With a time limit — adds a progress bar + overtime tracking
aegis start "coding" --limit 25m

# Detached mode — logs to DB and returns to shell
aegis start -d "coding"
```

### Track distractions

```bash
# Log a distraction urge (linked to active task if one is running)
aegis urge
```

In live TUI mode, press `u` to log an urge without leaving the session. At 3, 5, and 10 urges, a research paper highlight about dopamine and addiction is shown to help you stay focused.

### Stop a task

```bash
aegis stop
```

In live TUI mode, press `q` to stop and exit. A styled session summary card is shown with your duration and urge count.

### Check status

```bash
aegis status
```

### View stats

```bash
aegis stats          # last 7 days (default)
aegis stats day      # today
aegis stats week     # last 7 days with daily breakdown
aegis stats month    # current calendar month summary
aegis stats year     # current year with monthly breakdown
```

### View recent tasks

```bash
aegis history
```

Shows the 5 most recent completed tasks with duration, urge count, and timestamps in styled cards.

## Storage

All data is stored locally in `~/.aegis/aegis.db` (SQLite). No cloud, no accounts, no telemetry.

## Requirements

- macOS or Linux
- Go 1.24+ (only if installing with `go install`)
