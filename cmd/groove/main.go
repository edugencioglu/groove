package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/edugencioglu/groove/internal/claude"
	"github.com/edugencioglu/groove/internal/config"
	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/gitstatus"
	"github.com/edugencioglu/groove/internal/shell"
	"github.com/edugencioglu/groove/internal/tabs"
	"github.com/edugencioglu/groove/internal/tui"
)

const usage = `groove — Worktree Dashboard & Ghostty Tab Manager

Usage:
  groove [path]        Launch TUI dashboard (scans path, GROOVE_ROOT, or cwd)
  groove list [path]   List discovered projects and worktrees as text
  groove init zsh      Print shell integration snippet for .zshrc
  groove title         Print "Project/Worktree" label for current directory
  groove help          Show this help message

Environment:
  GROOVE_ROOT          Default root directory to scan for git repos
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	if len(args) == 0 {
		return cmdDashboard(nil)
	}

	switch args[0] {
	case "list":
		return cmdList(args[1:])
	case "init":
		return cmdInit(args[1:])
	case "title":
		return cmdTitle()
	case "help", "--help", "-h":
		fmt.Print(usage)
		return nil
	default:
		// If it looks like a flag, show help.
		if args[0][0] == '-' {
			fmt.Print(usage)
			return fmt.Errorf("unknown flag: %s", args[0])
		}
		// Treat as a path for the TUI.
		return cmdDashboard(args)
	}
}

func cmdList(args []string) error {
	root, err := resolveRoot(args)
	if err != nil {
		return err
	}

	ctx := context.Background()

	projects, err := discovery.Discover(ctx, root, discovery.DefaultGitRunner)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("No git repositories found.")
		return nil
	}

	// Enrich with git status (errors are non-fatal for listing).
	_ = gitstatus.EnrichWorktrees(ctx, projects, gitstatus.DefaultGitRunner)

	for _, p := range projects {
		fmt.Printf("%s (%s)\n", p.Name, p.Path)
		for _, wt := range p.Worktrees {
			status := "clean"
			if !wt.IsClean {
				status = fmt.Sprintf("%dM %dU", wt.Modified, wt.Untracked)
			}
			ab := ""
			if wt.Ahead > 0 || wt.Behind > 0 {
				ab = fmt.Sprintf(" ↑%d↓%d", wt.Ahead, wt.Behind)
			}
			fmt.Printf("  %-20s %-20s %s%s\n", wt.Name, wt.Branch, status, ab)
		}
	}

	return nil
}

func cmdInit(args []string) error {
	if len(args) == 0 || args[0] != "zsh" {
		return fmt.Errorf("usage: groove init zsh\n\nSupported shells: zsh")
	}
	fmt.Print(shell.ZshHook())
	return nil
}

func cmdTitle() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(shell.Title(dir))
	return nil
}

func cmdDashboard(args []string) error {
	if err := checkGit(); err != nil {
		return err
	}

	root, err := resolveRoot(args)
	if err != nil {
		return err
	}

	// Discovery function used for both initial load and refresh.
	discoverAndEnrich := func(ctx context.Context) ([]discovery.Project, error) {
		p, err := discovery.Discover(ctx, root, discovery.DefaultGitRunner)
		if err != nil {
			return nil, err
		}
		_ = gitstatus.EnrichWorktrees(ctx, p, gitstatus.DefaultGitRunner)
		tabs.DetectTabs(ctx, tabs.DefaultCmdRunner, p)
		claude.DetectClaude(ctx, tabs.DefaultCmdRunner, p)
		return p, nil
	}

	labels, err := config.LoadLabels()
	if err != nil {
		return fmt.Errorf("loading labels: %w", err)
	}

	m := tui.NewLoading(discoverAndEnrich).
		WithTabAction(ghosttyTabAction).
		WithRefreshFunc(discoverAndEnrich).
		WithLabels(labels).
		WithLabelSaveFn(config.SaveLabels)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// ghosttyTabAction creates or jumps to a Ghostty tab for the given worktree.
// If a tab is detected but the jump fails (title mismatch), it falls back to
// creating a new tab.
func ghosttyTabAction(wt discovery.Worktree, projectName string) error {
	if wt.TabCount > 0 {
		script := tabs.JumpTabScript(wt.Path)
		if err := runOsascript(script); err == nil {
			return nil
		}
		// Jump failed — fall back to new tab.
	}
	return runOsascript(tabs.NewTabScript(wt.Path))
}

// runOsascript executes an AppleScript and returns any error with stderr detail.
func runOsascript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

// resolveRoot determines the root directory to scan.
// Priority: positional arg > GROOVE_ROOT env var > current working directory.
func resolveRoot(args []string) (string, error) {
	var root string
	switch {
	case len(args) > 0:
		root = args[0]
	case os.Getenv("GROOVE_ROOT") != "":
		root = os.Getenv("GROOVE_ROOT")
	default:
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		root = cwd
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("cannot access %s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", root)
	}
	return abs, nil
}

// checkGit verifies that git is available on PATH.
func checkGit() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH — install git to use groove")
	}
	return nil
}
