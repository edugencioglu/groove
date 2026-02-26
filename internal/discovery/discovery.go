package discovery

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GitRunner executes a git command in the given directory and returns its
// combined stdout output. This function type enables dependency injection
// for testing (following the http.HandlerFunc pattern).
type GitRunner func(ctx context.Context, dir string, args ...string) (string, error)

// DefaultGitRunner runs git commands via os/exec.
func DefaultGitRunner(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %v in %s: %w", args, dir, err)
	}
	return string(out), nil
}

// Discover scans root for git repositories and returns a Project for each one,
// with worktrees populated from `git worktree list --porcelain`. Repos where
// git fails are skipped with a warning to stderr.
func Discover(ctx context.Context, root string, runner GitRunner) ([]Project, error) {
	repos, err := FindGitRepos(root)
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", root, err)
	}

	var projects []Project
	for _, repoPath := range repos {
		output, err := runner(ctx, repoPath, "worktree", "list", "--porcelain")
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", repoPath, err)
			continue
		}

		worktrees := ParseWorktreeList(output)
		if len(worktrees) == 0 {
			continue
		}

		projects = append(projects, Project{
			Name:      filepath.Base(repoPath),
			Path:      repoPath,
			Worktrees: worktrees,
		})
	}

	return projects, nil
}
