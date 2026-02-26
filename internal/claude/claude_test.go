package claude

import (
	"context"
	"fmt"
	"testing"

	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/tabs"
)

// ---------------------------------------------------------------------------
// FindClaudePIDs tests
// ---------------------------------------------------------------------------

func TestFindClaudePIDs(t *testing.T) {
	procs := []tabs.ProcessInfo{
		{PID: 1, PPID: 0, Comm: "/sbin/launchd"},
		{PID: 100, PPID: 1, Comm: "/usr/local/bin/claude"},
		{PID: 200, PPID: 1, Comm: "/bin/zsh"},
		{PID: 300, PPID: 1, Comm: "claude"},
		{PID: 400, PPID: 1, Comm: "/usr/bin/claude-code"},
	}

	pids := FindClaudePIDs(procs)
	if len(pids) != 3 {
		t.Fatalf("got %d PIDs, want 3: %v", len(pids), pids)
	}

	pidSet := map[int]bool{}
	for _, p := range pids {
		pidSet[p] = true
	}
	if !pidSet[100] || !pidSet[300] || !pidSet[400] {
		t.Errorf("expected PIDs 100, 300, 400, got %v", pids)
	}
}

func TestFindClaudePIDs_None(t *testing.T) {
	procs := []tabs.ProcessInfo{
		{PID: 1, PPID: 0, Comm: "/sbin/launchd"},
		{PID: 200, PPID: 1, Comm: "/bin/zsh"},
	}

	pids := FindClaudePIDs(procs)
	if len(pids) != 0 {
		t.Fatalf("got %d PIDs, want 0", len(pids))
	}
}

// ---------------------------------------------------------------------------
// ParsePgrepOutput tests
// ---------------------------------------------------------------------------

func TestParsePgrepOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"multiple PIDs", "12345\n67890\n", 2},
		{"single PID", "999\n", 1},
		{"empty", "", 0},
		{"with whitespace", "  123  \n  456  \n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pids := ParsePgrepOutput(tt.input)
			if len(pids) != tt.want {
				t.Errorf("got %d PIDs, want %d", len(pids), tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectClaude integration test (mock CmdRunner)
// ---------------------------------------------------------------------------

func TestDetectClaude_PSMatch(t *testing.T) {
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  100     1 /usr/local/bin/claude
  200     1 /bin/zsh
`

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		switch name {
		case "ps":
			return psOutput, nil
		case "/usr/sbin/lsof":
			for i, a := range args {
				if a == "-p" && i+1 < len(args) {
					if args[i+1] == "100" {
						return "p100\nfcwd\nn/home/user/project-a\n", nil
					}
				}
			}
			return "", fmt.Errorf("unknown pid")
		}
		return "", fmt.Errorf("unknown command: %s", name)
	}

	projects := []discovery.Project{
		{
			Name: "project-a",
			Path: "/home/user/project-a",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-a", Branch: "main"},
				{Name: "feature", Path: "/home/user/project-a-feature", Branch: "feature"},
			},
		},
	}

	DetectClaude(context.Background(), runner, projects)

	if !projects[0].Worktrees[0].HasClaude {
		t.Error("project-a/main should have HasClaude=true")
	}
	if projects[0].Worktrees[1].HasClaude {
		t.Error("project-a/feature should have HasClaude=false")
	}
}

func TestDetectClaude_PgrepFallback(t *testing.T) {
	// ps shows no "claude" in comm, but pgrep finds it (e.g. node-based)
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  100     1 /usr/bin/node
  200     1 /bin/zsh
`

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		switch name {
		case "ps":
			return psOutput, nil
		case "pgrep":
			return "100\n", nil
		case "/usr/sbin/lsof":
			for i, a := range args {
				if a == "-p" && i+1 < len(args) && args[i+1] == "100" {
					return "p100\nfcwd\nn/home/user/project-b\n", nil
				}
			}
			return "", fmt.Errorf("unknown pid")
		}
		return "", fmt.Errorf("unknown command: %s", name)
	}

	projects := []discovery.Project{
		{
			Name: "project-b",
			Path: "/home/user/project-b",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-b", Branch: "main"},
			},
		},
	}

	DetectClaude(context.Background(), runner, projects)

	if !projects[0].Worktrees[0].HasClaude {
		t.Error("project-b/main should have HasClaude=true via pgrep fallback")
	}
}

func TestDetectClaude_NoClaude(t *testing.T) {
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  200     1 /bin/zsh
`

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		switch name {
		case "ps":
			return psOutput, nil
		case "pgrep":
			return "", fmt.Errorf("no matches")
		}
		return "", fmt.Errorf("unknown command: %s", name)
	}

	projects := []discovery.Project{
		{
			Name: "project-a",
			Path: "/home/user/project-a",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-a", Branch: "main"},
			},
		},
	}

	DetectClaude(context.Background(), runner, projects)

	if projects[0].Worktrees[0].HasClaude {
		t.Error("should not detect claude when no claude processes exist")
	}
}

func TestDetectClaude_PSFails(t *testing.T) {
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		return "", fmt.Errorf("ps failed")
	}

	projects := []discovery.Project{
		{
			Name: "project-a",
			Path: "/home/user/project-a",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-a", Branch: "main"},
			},
		},
	}

	// Should not panic.
	DetectClaude(context.Background(), runner, projects)

	if projects[0].Worktrees[0].HasClaude {
		t.Error("should not detect claude when ps fails")
	}
}

func TestDetectClaude_LsofFails(t *testing.T) {
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  100     1 claude
`

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if name == "ps" {
			return psOutput, nil
		}
		return "", fmt.Errorf("lsof failed")
	}

	projects := []discovery.Project{
		{
			Name: "project-a",
			Path: "/home/user/project-a",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-a", Branch: "main"},
			},
		},
	}

	DetectClaude(context.Background(), runner, projects)

	if projects[0].Worktrees[0].HasClaude {
		t.Error("should not detect claude when lsof fails")
	}
}
