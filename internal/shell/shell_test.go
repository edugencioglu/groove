package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ZshHook tests
// ---------------------------------------------------------------------------

func TestZshHook(t *testing.T) {
	snippet := ZshHook()

	if !strings.Contains(snippet, "_groove_precmd()") {
		t.Error("snippet should define _groove_precmd function")
	}
	if !strings.Contains(snippet, "precmd_functions+=(_groove_precmd)") {
		t.Error("snippet should register precmd hook")
	}
	if !strings.Contains(snippet, "title") {
		t.Error("snippet should call title subcommand")
	}
	if !strings.Contains(snippet, `printf '\e]2;`) {
		t.Error("snippet should set terminal title via escape sequence")
	}
}

// ---------------------------------------------------------------------------
// Title tests
// ---------------------------------------------------------------------------

func TestTitle_MainRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "myproject")
	mkDir(t, repo, ".git")

	got := Title(repo)
	if got != "myproject" {
		t.Errorf("Title() = %q, want %q", got, "myproject")
	}
}

func TestTitle_LinkedWorktree(t *testing.T) {
	root := t.TempDir()

	// Set up main repo.
	mainRepo := filepath.Join(root, "myproject")
	mkDir(t, mainRepo, ".git", "worktrees", "feature-wt")

	// Set up linked worktree with .git file pointing to main repo.
	linkedWt := filepath.Join(root, "feature-wt")
	mkDir(t, linkedWt)
	gitdir := filepath.Join(mainRepo, ".git", "worktrees", "feature-wt")
	writeFile(t, filepath.Join(linkedWt, ".git"), "gitdir: "+gitdir+"\n")

	got := Title(linkedWt)
	if got != "myproject/feature-wt" {
		t.Errorf("Title() = %q, want %q", got, "myproject/feature-wt")
	}
}

func TestTitle_NotARepo(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "somedir")
	mkDir(t, dir)

	got := Title(dir)
	if got != "somedir" {
		t.Errorf("Title() = %q, want %q (basename fallback)", got, "somedir")
	}
}

func TestTitle_SubdirOfRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "myproject")
	mkDir(t, repo, ".git")
	subdir := filepath.Join(repo, "src", "pkg")
	mkDir(t, subdir)

	got := Title(subdir)
	if got != "myproject" {
		t.Errorf("Title() = %q, want %q (should find repo root)", got, "myproject")
	}
}

func TestTitle_SameNameWorktree(t *testing.T) {
	// Edge case: worktree dir has the same name as the project.
	root := t.TempDir()
	mainRepo := filepath.Join(root, "myproject")
	mkDir(t, mainRepo, ".git", "worktrees", "myproject")

	linkedWt := filepath.Join(root, "myproject-copy")
	// Rename to match project name for test.
	sameNameWt := filepath.Join(root, "myproject-samename")
	mkDir(t, sameNameWt)
	_ = linkedWt

	// Actually set up a worktree with the same name.
	wt := filepath.Join(root, "myproject2")
	mkDir(t, wt)
	gitdir := filepath.Join(mainRepo, ".git", "worktrees", "myproject2")
	mkDir(t, gitdir)
	writeFile(t, filepath.Join(wt, ".git"), "gitdir: "+gitdir+"\n")

	got := Title(wt)
	if got != "myproject/myproject2" {
		t.Errorf("Title() = %q, want %q", got, "myproject/myproject2")
	}
}

// ---------------------------------------------------------------------------
// findGitRoot tests
// ---------------------------------------------------------------------------

func TestFindGitRoot(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	mkDir(t, repo, ".git")
	deep := filepath.Join(repo, "a", "b", "c")
	mkDir(t, deep)

	got := findGitRoot(deep)
	if got != repo {
		t.Errorf("findGitRoot() = %q, want %q", got, repo)
	}
}

func TestFindGitRoot_NotFound(t *testing.T) {
	root := t.TempDir()
	got := findGitRoot(root)
	if got != "" {
		t.Errorf("findGitRoot() = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// resolveProjectName tests
// ---------------------------------------------------------------------------

func TestResolveProjectName(t *testing.T) {
	root := t.TempDir()
	gitFile := filepath.Join(root, "dotgit")
	writeFile(t, gitFile, "gitdir: /home/user/myproject/.git/worktrees/feature\n")

	got := resolveProjectName(gitFile)
	if got != "myproject" {
		t.Errorf("resolveProjectName() = %q, want %q", got, "myproject")
	}
}

func TestResolveProjectName_InvalidFormat(t *testing.T) {
	root := t.TempDir()
	gitFile := filepath.Join(root, "dotgit")
	writeFile(t, gitFile, "not a gitdir line\n")

	got := resolveProjectName(gitFile)
	if got != "" {
		t.Errorf("resolveProjectName() = %q, want empty", got)
	}
}

func TestResolveProjectName_NoWorktreesInPath(t *testing.T) {
	root := t.TempDir()
	gitFile := filepath.Join(root, "dotgit")
	writeFile(t, gitFile, "gitdir: /some/random/path\n")

	got := resolveProjectName(gitFile)
	if got != "" {
		t.Errorf("resolveProjectName() = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mkDir(t *testing.T, parts ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(parts...), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
