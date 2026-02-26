package discovery

import (
	"path/filepath"
	"strings"
)

// ParseWorktreeList parses the output of `git worktree list --porcelain` and
// returns a slice of Worktree structs. Each block in the porcelain output is
// separated by a blank line and contains lines like:
//
//	worktree /path/to/worktree
//	HEAD abc123
//	branch refs/heads/main
//
// Detached HEAD blocks have "detached" instead of "branch ...".
// Bare repos have "bare" instead of "branch ...".
func ParseWorktreeList(output string) []Worktree {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	blocks := splitBlocks(output)
	worktrees := make([]Worktree, 0, len(blocks))

	for _, block := range blocks {
		wt, ok := parseBlock(block)
		if ok {
			worktrees = append(worktrees, wt)
		}
	}

	return worktrees
}

// splitBlocks splits porcelain output into blocks separated by blank lines.
func splitBlocks(output string) []string {
	var blocks []string
	var current strings.Builder

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			if current.Len() > 0 {
				blocks = append(blocks, current.String())
				current.Reset()
			}
			continue
		}
		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)
	}
	if current.Len() > 0 {
		blocks = append(blocks, current.String())
	}

	return blocks
}

// parseBlock parses a single porcelain block into a Worktree.
func parseBlock(block string) (Worktree, bool) {
	var wt Worktree

	for _, line := range strings.Split(block, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wt.Path = strings.TrimPrefix(line, "worktree ")
			wt.Name = filepath.Base(wt.Path)
		case strings.HasPrefix(line, "branch "):
			// branch refs/heads/feature/auth -> feature/auth
			ref := strings.TrimPrefix(line, "branch ")
			wt.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			wt.Branch = "(detached)"
		case line == "bare":
			wt.Branch = "(bare)"
		}
	}

	if wt.Path == "" {
		return Worktree{}, false
	}
	// Skip bare worktree entries — they're not real working directories.
	if wt.Branch == "(bare)" {
		return Worktree{}, false
	}
	return wt, true
}
