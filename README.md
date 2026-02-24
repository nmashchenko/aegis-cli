# aegis-cli
<img width="1103" height="245" alt="Screenshot 2026-02-24 at 5 21 27 PM" src="https://github.com/user-attachments/assets/e059598a-ca90-4870-9946-48fc8b32e7c4" />

A lightweight CLI tool for tracking deep work sessions and distraction impulses. Built in Go, runs locally, no external services.

## Install

```bash
go install github.com/nmashchenko/aegis-cli@latest
```

Or build from source:

```bash
git clone https://github.com/nmashchenko/aegis-cli.git
cd aegis-cli
go build -o aegis .
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

In live TUI mode, press `u` to log an urge without leaving the session.

### Stop a task

```bash
aegis stop
```

In live TUI mode, press `q` to stop and exit.

### Check status

```bash
aegis status
```

### View stats

```bash
aegis stats          # today (default)
aegis stats today
aegis stats week     # last 7 days
aegis stats month    # last 30 days
aegis stats year     # last 365 days
```

### View recent tasks

```bash
aegis history
```

Shows the 5 most recent completed tasks with duration, urge count, and timestamps in styled cards.

## Storage

All data is stored locally in `~/.aegis/aegis.db` (SQLite). No cloud, no accounts, no telemetry.

## Requirements

- macOS (designed for, but should work on Linux)
- Go 1.21+
