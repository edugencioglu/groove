package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// skipDirs contains directory names that should never be recursed into.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".cache":       true,
	".venv":        true,
	"__pycache__":  true,
	".terraform":   true,
	".build":       true,
}

// FindGitRepos walks root and returns absolute paths of directories that
// contain a .git directory (not a .git file, which indicates a linked worktree).
// It stops recursing into a directory once a .git dir is found.
func FindGitRepos(root string) ([]string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var repos []string

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		name := d.Name()

		// Skip known uninteresting directories.
		if skipDirs[name] && path != root {
			return filepath.SkipDir
		}

		// Check for .git directory (not file) inside this directory.
		gitPath := filepath.Join(path, ".git")
		info, err := os.Lstat(gitPath)
		if err != nil {
			return nil // no .git here, keep walking
		}

		if info.IsDir() {
			repos = append(repos, path)
			return filepath.SkipDir // don't recurse into repos
		}

		// .git is a file — could be a linked worktree or a bare-repo root.
		// Read it to check: "gitdir: ./.bare" (or similar local bare dir)
		// means this is the root of a bare-repo worktree setup.
		if isLocalBareGitFile(gitPath) {
			repos = append(repos, path)
		}
		return filepath.SkipDir
	})

	return repos, err
}

// isLocalBareGitFile returns true if the .git file points to a local bare repo
// directory (e.g., "gitdir: ./.bare" or "gitdir: .bare"). This distinguishes
// bare-repo worktree roots from linked worktrees (which point to absolute paths
// like "../.bare/worktrees/feature").
func isLocalBareGitFile(gitPath string) bool {
	data, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}
	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "gitdir: ") {
		return false
	}
	target := strings.TrimPrefix(content, "gitdir: ")
	// Local bare repos use relative paths like ".bare" or "./.bare".
	// Linked worktrees use paths containing "/worktrees/".
	return !strings.Contains(target, "/worktrees/")
}
