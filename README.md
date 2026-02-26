<h1 align="center">
  <br>
  🌳
  <br>
  groove
</h1>
<p align="center">Worktree Dashboard & Ghostty Tab Manager</p>

A TUI dashboard for managing git worktrees and [Ghostty](https://ghostty.org) tabs on macOS.

Scans a directory for git repos, shows their worktree status at a glance, and lets you create or jump to Ghostty terminal tabs — one per worktree.

## Features

- Auto-discovers git repositories and their worktrees
- Shows branch, dirty/clean status, ahead/behind counts
- Detects active Ghostty tabs and Claude Code sessions per worktree
- Create or jump to Ghostty tabs with Enter
- Async disk usage display per worktree
- Auto-refresh every 30 seconds
- Responsive two-column grid layout
- Shell integration for terminal tab titles

## Install

```bash
go install github.com/edugencioglu/groove/cmd/groove@latest
```

Or build from source:

```bash
git clone https://github.com/edugencioglu/groove.git
cd groove
go build ./cmd/groove/
```

## Usage

```bash
groove                    # Launch dashboard (scans GROOVE_ROOT or cwd)
groove ~/dev              # Launch dashboard scanning ~/dev
groove list [path]        # List projects and worktrees as plain text
groove init zsh           # Print shell integration snippet
groove title              # Print "Project/Worktree" label for current dir
```

### Shell Integration

Add to your `.zshrc` to automatically set tab titles:

```bash
eval "$(groove init zsh)"
```

### Environment Variables

- `GROOVE_ROOT` — Default root directory to scan for git repos

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Tab` / `Shift+Tab` | Switch between projects |
| `Enter` | Open or jump to Ghostty tab |
| `n` | Rename / label selected worktree |
| `r` | Refresh |
| `q` | Quit |

## Configuration

Labels and settings are stored in `~/.config/groove/`.

- **labels.json** — Custom labels assigned to worktrees via `n` key

## Requirements

- macOS (uses AppleScript for Ghostty tab management)
- [Git](https://git-scm.com)
- [Ghostty](https://ghostty.org) (optional — tab features degrade gracefully)
