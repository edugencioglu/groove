# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`groove` is a Go TUI application — a worktree dashboard and Ghostty tab manager for macOS. It auto-discovers git worktrees, shows their status, detects active Ghostty tabs and Claude Code sessions, and lets you create/jump to Ghostty tabs via AppleScript.

## Tech Stack

- **Go** with **Bubbletea** (Elm-architecture TUI), **Lipgloss** (styling), **Bubbles** (components)
- AppleScript via `os/exec` for Ghostty tab control (macOS only)
- No external dependencies beyond Charm libraries — git ops, process scanning, and AppleScript all use `os/exec`

## Build & Test Commands

```bash
go build ./cmd/groove/           # Build the binary
go test ./...                    # Run all tests
go test ./internal/discovery/    # Run tests for a single package
go test -race ./...              # Tests with race detector
go test -run TestName ./...      # Run a single test by name
go run ./cmd/groove/             # Run directly
go vet ./...                     # Static analysis
go test ./internal/tui/ -update  # Update golden files after intentional TUI changes
```

## Architecture

```
cmd/groove/main.go              # Entry point, CLI subcommand routing
internal/
  discovery/                    # Scan dirs for git repos, parse `git worktree list --porcelain`
  gitstatus/                    # Branch name, dirty/clean, ahead/behind per worktree
  tabs/                         # Ghostty tab detection (process scanning) and actions (AppleScript)
  claude/                       # Detect active Claude Code sessions via process cwd matching
  diskusage/                    # Async `du -sh` per worktree, delivered via Bubbletea Cmd/Msg
  tui/                          # Bubbletea model/update/view + Lipgloss styles
  shell/                        # Shell integration: `groove init zsh` hook, `groove title` label
  config/                       # Labels persistence (~/.config/groove/labels.json)
  theme/                        # Color themes (10 preset palettes) with card/accent/state colors
```

## Key Architectural Patterns

### Function-Type Dependency Injection

Every package that shells out defines its own runner type for testability:

| Package    | Type         | Signature                                          |
|------------|-------------|----------------------------------------------------|
| discovery  | `GitRunner` | `func(ctx, dir, args ...string) (string, error)`  |
| gitstatus  | `GitRunner` | Same as above                                      |
| tabs       | `CmdRunner` | `func(ctx, name, args ...string) (string, error)` |
| diskusage  | `CmdRunner` | Same as tabs                                       |

Each package provides a `Default*Runner` that calls `os/exec`. Tests inject mock runners that return hardcoded output, switching on the directory or command name. The `claude` package reuses `tabs.CmdRunner`.

### Data Enrichment Pipeline

`discovery.Project` and `discovery.Worktree` are populated in stages — each step mutates in-place:
1. `discovery.Discover()` → base project/worktree structure
2. `gitstatus.EnrichWorktrees()` → branch, dirty/clean, ahead/behind
3. `tabs.DetectTabs()` → Ghostty tab counts
4. `claude.DetectClaude()` → Claude Code session detection
5. `diskusage.FetchAll()` → async disk usage via Bubbletea messages

### TUI Builder Pattern

The model uses method chaining for configuration:
```go
tui.NewLoading(discoverFn).
    WithTabAction(fn).
    WithRefreshFunc(fn).
    WithLabels(labels).
    WithLabelSaveFn(config.SaveLabels)
```

The view has three layers: splash screen (loading) → dashboard (header + tab bar + cards grid) → modal overlay (rename).

### Concurrency

- `gitstatus.EnrichWorktrees`: goroutine per worktree with WaitGroup + Mutex
- `diskusage.FetchAll`: returns `[]tea.Cmd` batched by Bubbletea
- `groove title` avoids git binary entirely — walks filesystem to find `.git` for <10ms latency

## Testing Patterns

- **Table-driven tests** for all parsing functions (porcelain output, git status, ahead/behind)
- **Mock runners** inject fake command output — tests never call real git/ps/lsof
- **`t.TempDir()` + real dirs** for filesystem scanning tests (`FindGitRepos`)
- **teatest golden files** for TUI snapshots at fixed 80x24 terminal size. The `noDuModel()` helper creates deterministic models without async disk usage.
- **`WT_CONFIG_DIR` env var** overrides config directory in tests (`t.Setenv`)

## Key Design Decisions

- **Concurrency everywhere:** All git commands, process scanning, and disk usage run concurrently. The TUI must never freeze.
- **`groove title` must be <10ms** since it runs on every zsh prompt via precmd hook.
- **Process detection on macOS:** Use `ps` output parsing and `lsof -a -p <pid> -d cwd -Fn` for cwd resolution. No `/proc` filesystem on macOS.
- **Graceful degradation:** If Ghostty isn't running, tab features show "N/A". If git isn't found, show a clear error. Errors in enrichment steps are non-fatal.
- **Grid TUI layout:** Two-column grid (one per project), rows for worktrees. Arrow keys navigate, Enter opens/jumps to Ghostty tab.
- **Worktree naming:** Derived from directory name, not branch name.
- **AppleScript for Ghostty:** Uses `System Events` keystroke simulation for new tabs and Accessibility API for tab switching.
