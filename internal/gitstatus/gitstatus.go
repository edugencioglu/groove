package gitstatus

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/edugencioglu/groove/internal/discovery"
)

// GitRunner executes a git command in the given directory and returns its output.
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

// ParseStatus parses `git status --porcelain` output and returns counts of
// modified and untracked files.
func ParseStatus(output string) (modified, untracked int) {
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 2 {
			continue
		}
		xy := line[:2]
		switch {
		case xy == "??":
			untracked++
		default:
			// Any other non-empty status indicator means modified/added/deleted/etc.
			modified++
		}
	}
	return modified, untracked
}

// ParseAheadBehind parses the output of
// `git rev-list --left-right --count HEAD...@{upstream}` which produces
// a line like "3\t1" meaning 3 ahead, 1 behind.
func ParseAheadBehind(output string) (ahead, behind int) {
	output = strings.TrimSpace(output)
	if output == "" {
		return 0, 0
	}
	parts := strings.Split(output, "\t")
	if len(parts) != 2 {
		return 0, 0
	}
	ahead, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
	behind, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
	return ahead, behind
}

// EnrichWorktrees runs git status and ahead/behind checks concurrently for
// all worktrees across all projects, updating the worktree fields in place.
func EnrichWorktrees(ctx context.Context, projects []discovery.Project, runner GitRunner) error {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	for i := range projects {
		for j := range projects[i].Worktrees {
			wt := &projects[i].Worktrees[j]
			wg.Add(1)
			go func(wt *discovery.Worktree) {
				defer wg.Done()
				if err := enrichOne(ctx, wt, runner); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				}
			}(wt)
		}
	}

	wg.Wait()
	return firstErr
}

func enrichOne(ctx context.Context, wt *discovery.Worktree, runner GitRunner) error {
	// Run status and ahead/behind concurrently for this worktree.
	var statusOut, abOut string
	var statusErr, abErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		statusOut, statusErr = runner(ctx, wt.Path, "status", "--porcelain")
	}()

	go func() {
		defer wg.Done()
		abOut, abErr = runner(ctx, wt.Path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	}()

	wg.Wait()

	if statusErr != nil {
		return fmt.Errorf("git status in %s: %w", wt.Path, statusErr)
	}

	wt.Modified, wt.Untracked = ParseStatus(statusOut)
	wt.IsClean = wt.Modified == 0 && wt.Untracked == 0

	// Ahead/behind may fail if there's no upstream — that's OK.
	if abErr == nil {
		wt.Ahead, wt.Behind = ParseAheadBehind(abOut)
	}

	return nil
}
