package gitstatus

import (
	"context"
	"fmt"
	"testing"

	"github.com/edugencioglu/groove/internal/discovery"
)

// ---------------------------------------------------------------------------
// ParseStatus tests (table-driven)
// ---------------------------------------------------------------------------

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantModified  int
		wantUntracked int
	}{
		{
			name:          "clean repo",
			input:         "",
			wantModified:  0,
			wantUntracked: 0,
		},
		{
			name:          "one modified file",
			input:         " M file.go\n",
			wantModified:  1,
			wantUntracked: 0,
		},
		{
			name:          "one untracked file",
			input:         "?? newfile.go\n",
			wantModified:  0,
			wantUntracked: 1,
		},
		{
			name:          "mixed changes",
			input:         " M file1.go\nM  file2.go\n?? new.go\n?? another.go\nA  added.go\n",
			wantModified:  3,
			wantUntracked: 2,
		},
		{
			name:          "deleted file",
			input:         " D removed.go\n",
			wantModified:  1,
			wantUntracked: 0,
		},
		{
			name:          "renamed file",
			input:         "R  old.go -> new.go\n",
			wantModified:  1,
			wantUntracked: 0,
		},
		{
			name:          "staged and unstaged",
			input:         "MM file.go\n",
			wantModified:  1,
			wantUntracked: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, unt := ParseStatus(tt.input)
			if mod != tt.wantModified {
				t.Errorf("modified = %d, want %d", mod, tt.wantModified)
			}
			if unt != tt.wantUntracked {
				t.Errorf("untracked = %d, want %d", unt, tt.wantUntracked)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseAheadBehind tests (table-driven)
// ---------------------------------------------------------------------------

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAhead  int
		wantBehind int
	}{
		{
			name:       "ahead and behind",
			input:      "3\t1\n",
			wantAhead:  3,
			wantBehind: 1,
		},
		{
			name:       "only ahead",
			input:      "5\t0\n",
			wantAhead:  5,
			wantBehind: 0,
		},
		{
			name:       "only behind",
			input:      "0\t2\n",
			wantAhead:  0,
			wantBehind: 2,
		},
		{
			name:       "in sync",
			input:      "0\t0\n",
			wantAhead:  0,
			wantBehind: 0,
		},
		{
			name:       "empty output",
			input:      "",
			wantAhead:  0,
			wantBehind: 0,
		},
		{
			name:       "malformed output",
			input:      "garbage",
			wantAhead:  0,
			wantBehind: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind := ParseAheadBehind(tt.input)
			if ahead != tt.wantAhead {
				t.Errorf("ahead = %d, want %d", ahead, tt.wantAhead)
			}
			if behind != tt.wantBehind {
				t.Errorf("behind = %d, want %d", behind, tt.wantBehind)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// EnrichWorktrees integration test (mock GitRunner)
// ---------------------------------------------------------------------------

func TestEnrichWorktrees(t *testing.T) {
	projects := []discovery.Project{
		{
			Name: "project-a",
			Path: "/tmp/project-a",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/tmp/project-a", Branch: "main"},
				{Name: "feature", Path: "/tmp/project-a-feature", Branch: "feature/auth"},
			},
		},
		{
			Name: "project-b",
			Path: "/tmp/project-b",
			Worktrees: []discovery.Worktree{
				{Name: "project-b", Path: "/tmp/project-b", Branch: "main"},
			},
		},
	}

	runner := func(_ context.Context, dir string, args ...string) (string, error) {
		cmd := args[0]
		switch {
		case cmd == "status" && dir == "/tmp/project-a":
			return "", nil // clean
		case cmd == "status" && dir == "/tmp/project-a-feature":
			return " M file1.go\n M file2.go\n?? new.go\n", nil // 2 modified, 1 untracked
		case cmd == "status" && dir == "/tmp/project-b":
			return "?? todo.txt\n", nil // 1 untracked
		case cmd == "rev-list" && dir == "/tmp/project-a":
			return "0\t0\n", nil
		case cmd == "rev-list" && dir == "/tmp/project-a-feature":
			return "2\t1\n", nil
		case cmd == "rev-list" && dir == "/tmp/project-b":
			return "", fmt.Errorf("no upstream") // no upstream configured
		default:
			return "", fmt.Errorf("unexpected: %s %v", dir, args)
		}
	}

	err := EnrichWorktrees(context.Background(), projects, runner)
	if err != nil {
		t.Fatal(err)
	}

	// project-a main: clean, 0/0
	wt := projects[0].Worktrees[0]
	if !wt.IsClean {
		t.Error("project-a/main should be clean")
	}
	if wt.Modified != 0 || wt.Untracked != 0 {
		t.Errorf("project-a/main: modified=%d untracked=%d, want 0/0", wt.Modified, wt.Untracked)
	}

	// project-a feature: dirty, 2M 1U, 2 ahead 1 behind
	wt = projects[0].Worktrees[1]
	if wt.IsClean {
		t.Error("project-a/feature should be dirty")
	}
	if wt.Modified != 2 {
		t.Errorf("project-a/feature: modified=%d, want 2", wt.Modified)
	}
	if wt.Untracked != 1 {
		t.Errorf("project-a/feature: untracked=%d, want 1", wt.Untracked)
	}
	if wt.Ahead != 2 || wt.Behind != 1 {
		t.Errorf("project-a/feature: ahead=%d behind=%d, want 2/1", wt.Ahead, wt.Behind)
	}

	// project-b: 1 untracked, no upstream (ahead/behind stay 0)
	wt = projects[1].Worktrees[0]
	if wt.IsClean {
		t.Error("project-b should be dirty")
	}
	if wt.Untracked != 1 {
		t.Errorf("project-b: untracked=%d, want 1", wt.Untracked)
	}
	if wt.Ahead != 0 || wt.Behind != 0 {
		t.Errorf("project-b: ahead=%d behind=%d, want 0/0 (no upstream)", wt.Ahead, wt.Behind)
	}
}

func TestEnrichWorktrees_StatusError(t *testing.T) {
	projects := []discovery.Project{
		{
			Name: "broken",
			Path: "/tmp/broken",
			Worktrees: []discovery.Worktree{
				{Name: "broken", Path: "/tmp/broken", Branch: "main"},
			},
		},
	}

	runner := func(_ context.Context, dir string, args ...string) (string, error) {
		if args[0] == "status" {
			return "", fmt.Errorf("not a git repo")
		}
		return "0\t0\n", nil
	}

	err := EnrichWorktrees(context.Background(), projects, runner)
	if err == nil {
		t.Error("expected error when git status fails")
	}
}
