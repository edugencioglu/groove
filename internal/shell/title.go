package shell

import (
	"os"
	"path/filepath"
	"strings"
)

// Title returns a label for the current directory suitable for a tab title.
// If the directory is inside a git worktree, it returns "Project/Worktree".
// Otherwise it returns the basename of the directory.
//
// This must be fast (<10ms) since it runs on every shell prompt. It avoids
// shelling out to git and instead reads filesystem markers directly.
func Title(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return filepath.Base(dir)
	}

	// Walk up to find the git root (directory containing .git dir or file).
	gitRoot := findGitRoot(dir)
	if gitRoot == "" {
		return filepath.Base(dir)
	}

	// Check if this is a linked worktree (.git is a file) or a main repo (.git is a dir).
	gitPath := filepath.Join(gitRoot, ".git")
	info, err := os.Lstat(gitPath)
	if err != nil {
		return filepath.Base(dir)
	}

	worktreeName := filepath.Base(gitRoot)

	if info.IsDir() {
		// Main repo — the project name IS the worktree name.
		return worktreeName
	}

	// Linked worktree — .git is a file containing "gitdir: /path/to/.git/worktrees/<name>".
	// The main repo is the parent of the .git directory referenced.
	projectName := resolveProjectName(gitPath)
	if projectName == "" {
		return worktreeName
	}

	if projectName == worktreeName {
		return worktreeName
	}
	return projectName + "/" + worktreeName
}

// findGitRoot walks up from dir looking for a directory containing .git.
func findGitRoot(dir string) string {
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Lstat(gitPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // reached filesystem root
		}
		dir = parent
	}
}

// resolveProjectName reads a .git file (linked worktree) and resolves the
// main repository's project name. The file contains a line like:
//
//	gitdir: /path/to/main-repo/.git/worktrees/wt-name
//
// We extract the main repo path and return its basename.
func resolveProjectName(gitFilePath string) string {
	data, err := os.ReadFile(gitFilePath)
	if err != nil {
		return ""
	}

	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir: ") {
		return ""
	}

	gitdir := strings.TrimPrefix(line, "gitdir: ")
	// gitdir is typically: /path/to/main-repo/.git/worktrees/<name>
	// We want: /path/to/main-repo
	// Walk up past "worktrees/<name>" and ".git".
	if i := strings.Index(gitdir, "/.git/worktrees/"); i >= 0 {
		mainRepo := gitdir[:i]
		return filepath.Base(mainRepo)
	}

	return ""
}
