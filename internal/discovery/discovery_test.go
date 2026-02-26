package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// ParseWorktreeList tests (table-driven)
// ---------------------------------------------------------------------------

func TestParseWorktreeList(t *testing.T) {
	tests := []struct {
		name string
		input string
		want  []Worktree
	}{
		{
			name: "single worktree",
			input: "worktree /home/user/project\nHEAD abc123\nbranch refs/heads/main\n",
			want: []Worktree{
				{Name: "project", Path: "/home/user/project", Branch: "main"},
			},
		},
		{
			name: "multiple worktrees",
			input: `worktree /home/user/project
HEAD abc123
branch refs/heads/main

worktree /home/user/project-feature
HEAD def456
branch refs/heads/feature
`,
			want: []Worktree{
				{Name: "project", Path: "/home/user/project", Branch: "main"},
				{Name: "project-feature", Path: "/home/user/project-feature", Branch: "feature"},
			},
		},
		{
			name: "branch with slashes",
			input: "worktree /home/user/project\nHEAD abc123\nbranch refs/heads/feature/auth/oauth\n",
			want: []Worktree{
				{Name: "project", Path: "/home/user/project", Branch: "feature/auth/oauth"},
			},
		},
		{
			name: "detached HEAD",
			input: "worktree /home/user/project\nHEAD abc123\ndetached\n",
			want: []Worktree{
				{Name: "project", Path: "/home/user/project", Branch: "(detached)"},
			},
		},
		{
			name:  "bare repo entry skipped",
			input: "worktree /home/user/project.git\nHEAD abc123\nbare\n",
			want:  nil,
		},
		{
			name: "bare entry skipped but worktrees kept",
			input: `worktree /home/user/project/.bare
bare

worktree /home/user/project/main
HEAD abc123
branch refs/heads/main
`,
			want: []Worktree{
				{Name: "main", Path: "/home/user/project/main", Branch: "main"},
			},
		},
		{
			name:  "empty output",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   \n\n  \n",
			want:  nil,
		},
		{
			name: "trailing newlines",
			input: "worktree /home/user/project\nHEAD abc123\nbranch refs/heads/main\n\n\n",
			want: []Worktree{
				{Name: "project", Path: "/home/user/project", Branch: "main"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseWorktreeList(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d worktrees, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("worktree[%d].Name = %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Path != tt.want[i].Path {
					t.Errorf("worktree[%d].Path = %q, want %q", i, got[i].Path, tt.want[i].Path)
				}
				if got[i].Branch != tt.want[i].Branch {
					t.Errorf("worktree[%d].Branch = %q, want %q", i, got[i].Branch, tt.want[i].Branch)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FindGitRepos tests (using t.TempDir)
// ---------------------------------------------------------------------------

// mkDir creates a directory (and parents) under root.
func mkDir(t *testing.T, parts ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(parts...), 0o755); err != nil {
		t.Fatal(err)
	}
}

// mkFile creates a file under root.
func mkFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindGitRepos_SingleRepo(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "myrepo", ".git")

	repos, err := FindGitRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(repos))
	}
	if filepath.Base(repos[0]) != "myrepo" {
		t.Errorf("got %q, want myrepo", filepath.Base(repos[0]))
	}
}

func TestFindGitRepos_MultipleRepos(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "repo-a", ".git")
	mkDir(t, root, "repo-b", ".git")
	mkDir(t, root, "subdir", "repo-c", ".git")

	repos, err := FindGitRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 3 {
		t.Fatalf("got %d repos, want 3", len(repos))
	}
}

func TestFindGitRepos_SkipsNodeModules(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "myrepo", ".git")
	mkDir(t, root, "node_modules", "dep", ".git")

	repos, err := FindGitRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1 (should skip node_modules)", len(repos))
	}
}

func TestFindGitRepos_SkipsLinkedWorktreeGitFile(t *testing.T) {
	root := t.TempDir()
	// A linked worktree has a .git *file* pointing to a /worktrees/ path
	mkDir(t, root, "linked-wt")
	mkFile(t, filepath.Join(root, "linked-wt", ".git"), "gitdir: /somewhere/else/.bare/worktrees/linked-wt")

	repos, err := FindGitRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Fatalf("got %d repos, want 0 (should skip linked worktree .git file)", len(repos))
	}
}

func TestFindGitRepos_BareRepoGitFile(t *testing.T) {
	root := t.TempDir()
	// A bare-repo root has a .git *file* pointing to a local .bare dir
	mkDir(t, root, "myproject", ".bare")
	mkFile(t, filepath.Join(root, "myproject", ".git"), "gitdir: ./.bare")

	repos, err := FindGitRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1 (should detect bare-repo .git file)", len(repos))
	}
	if filepath.Base(repos[0]) != "myproject" {
		t.Errorf("got %q, want myproject", filepath.Base(repos[0]))
	}
}

func TestFindGitRepos_NestedRepos(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "outer", ".git")
	mkDir(t, root, "outer", "inner", ".git") // should not be found

	repos, err := FindGitRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1 (should only find outer)", len(repos))
	}
	if filepath.Base(repos[0]) != "outer" {
		t.Errorf("got %q, want outer", filepath.Base(repos[0]))
	}
}

// ---------------------------------------------------------------------------
// Discover integration test (mock GitRunner)
// ---------------------------------------------------------------------------

func TestDiscover(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "project-a", ".git")
	mkDir(t, root, "project-b", ".git")

	porcelainA := "worktree %s\nHEAD abc123\nbranch refs/heads/main\n"
	porcelainB := "worktree %s\nHEAD def456\nbranch refs/heads/develop\n\nworktree %s\nHEAD 789abc\nbranch refs/heads/feature/login\n"

	pathA := filepath.Join(root, "project-a")
	pathB := filepath.Join(root, "project-b")
	pathBwt := filepath.Join(root, "project-b-login")

	runner := func(_ context.Context, dir string, args ...string) (string, error) {
		switch dir {
		case pathA:
			return fmt.Sprintf(porcelainA, pathA), nil
		case pathB:
			return fmt.Sprintf(porcelainB, pathB, pathBwt), nil
		default:
			return "", fmt.Errorf("unexpected dir: %s", dir)
		}
	}

	projects, err := Discover(context.Background(), root, runner)
	if err != nil {
		t.Fatal(err)
	}

	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}

	// project-a has 1 worktree
	if len(projects[0].Worktrees) != 1 {
		t.Errorf("project-a: got %d worktrees, want 1", len(projects[0].Worktrees))
	}

	// project-b has 2 worktrees
	if len(projects[1].Worktrees) != 2 {
		t.Errorf("project-b: got %d worktrees, want 2", len(projects[1].Worktrees))
	}
	if projects[1].Worktrees[1].Branch != "feature/login" {
		t.Errorf("project-b wt2 branch = %q, want feature/login", projects[1].Worktrees[1].Branch)
	}
}

func TestDiscover_GitFailureSkipsRepo(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "good-repo", ".git")
	mkDir(t, root, "bad-repo", ".git")

	goodPath := filepath.Join(root, "good-repo")

	runner := func(_ context.Context, dir string, args ...string) (string, error) {
		if filepath.Base(dir) == "bad-repo" {
			return "", fmt.Errorf("git failed")
		}
		return fmt.Sprintf("worktree %s\nHEAD abc\nbranch refs/heads/main\n", dir), nil
	}

	projects, err := Discover(context.Background(), root, runner)
	if err != nil {
		t.Fatal(err)
	}

	if len(projects) != 1 {
		t.Fatalf("got %d projects, want 1 (bad-repo should be skipped)", len(projects))
	}
	if projects[0].Path != goodPath {
		t.Errorf("got path %q, want %q", projects[0].Path, goodPath)
	}
}
