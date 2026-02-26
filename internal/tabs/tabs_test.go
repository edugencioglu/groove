package tabs

import (
	"context"
	"fmt"
	"testing"

	"github.com/edugencioglu/groove/internal/discovery"
)

// ---------------------------------------------------------------------------
// ParsePSOutput tests
// ---------------------------------------------------------------------------

func TestParsePSOutput(t *testing.T) {
	input := `  PID  PPID COMM
    1     0 /sbin/launchd
  501     1 /Applications/Ghostty.app/Contents/MacOS/ghostty
  502   501 /usr/bin/login
  503   502 -zsh
  504   503 /usr/local/bin/claude
  600     1 /Applications/Safari.app/Contents/MacOS/Safari
`

	procs := ParsePSOutput(input)

	if len(procs) != 6 {
		t.Fatalf("got %d procs, want 6", len(procs))
	}
	if procs[0].PID != 1 || procs[0].PPID != 0 {
		t.Errorf("proc[0] = %+v, want PID=1 PPID=0", procs[0])
	}
	if procs[1].PID != 501 {
		t.Errorf("proc[1].PID = %d, want 501", procs[1].PID)
	}
}

func TestParsePSOutput_Empty(t *testing.T) {
	procs := ParsePSOutput("")
	if len(procs) != 0 {
		t.Fatalf("got %d procs, want 0", len(procs))
	}
}

// ---------------------------------------------------------------------------
// FindGhosttyShellPIDs tests
// ---------------------------------------------------------------------------

func TestFindGhosttyShellPIDs(t *testing.T) {
	procs := []ProcessInfo{
		{PID: 1, PPID: 0, Comm: "/sbin/launchd"},
		{PID: 501, PPID: 1, Comm: "/Applications/Ghostty.app/Contents/MacOS/ghostty"},
		{PID: 502, PPID: 501, Comm: "/usr/bin/login"},
		{PID: 503, PPID: 502, Comm: "-zsh"},
		{PID: 504, PPID: 501, Comm: "/usr/bin/login"},
		{PID: 505, PPID: 504, Comm: "-zsh"},
		{PID: 600, PPID: 1, Comm: "/bin/zsh"}, // not a child of ghostty
	}

	pids := FindGhosttyShellPIDs(procs)

	if len(pids) != 2 {
		t.Fatalf("got %d shell PIDs, want 2: %v", len(pids), pids)
	}

	pidSet := map[int]bool{}
	for _, p := range pids {
		pidSet[p] = true
	}
	if !pidSet[503] || !pidSet[505] {
		t.Errorf("expected PIDs 503 and 505, got %v", pids)
	}
}

func TestFindGhosttyShellPIDs_NoGhostty(t *testing.T) {
	procs := []ProcessInfo{
		{PID: 1, PPID: 0, Comm: "/sbin/launchd"},
		{PID: 100, PPID: 1, Comm: "/bin/zsh"},
	}

	pids := FindGhosttyShellPIDs(procs)
	if len(pids) != 0 {
		t.Fatalf("got %d shell PIDs, want 0 (no ghostty)", len(pids))
	}
}

func TestFindGhosttyShellPIDs_BashAndFish(t *testing.T) {
	procs := []ProcessInfo{
		{PID: 1, PPID: 0, Comm: "/sbin/launchd"},
		{PID: 501, PPID: 1, Comm: "ghostty"},
		{PID: 502, PPID: 501, Comm: "-bash"},
		{PID: 503, PPID: 501, Comm: "fish"},
	}

	pids := FindGhosttyShellPIDs(procs)
	if len(pids) != 2 {
		t.Fatalf("got %d shell PIDs, want 2: %v", len(pids), pids)
	}
}

// ---------------------------------------------------------------------------
// ParseLsofCwd tests
// ---------------------------------------------------------------------------

func TestParseLsofCwd(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal output",
			input: "p12345\nfcwd\nn/Users/emre/project\n",
			want:  "/Users/emre/project",
		},
		{
			name:  "empty output",
			input: "",
			want:  "",
		},
		{
			name:  "no cwd line",
			input: "p12345\nfcwd\n",
			want:  "",
		},
		{
			name:  "path with spaces",
			input: "p999\nfcwd\nn/Users/emre/my project/src\n",
			want:  "/Users/emre/my project/src",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLsofCwd(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectTabs integration test (mock CmdRunner)
// ---------------------------------------------------------------------------

func TestDetectTabs(t *testing.T) {
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  501     1 /Applications/Ghostty.app/Contents/MacOS/ghostty
  502   501 /usr/bin/login
  503   502 -zsh
  504   501 /usr/bin/login
  505   504 -zsh
`

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		switch name {
		case "ps":
			return psOutput, nil
		case "/usr/sbin/lsof":
			// Find the -p arg to determine which PID was queried.
			for i, a := range args {
				if a == "-p" && i+1 < len(args) {
					switch args[i+1] {
					case "503":
						return "p503\nfcwd\nn/home/user/project-a\n", nil
					case "505":
						return "p505\nfcwd\nn/home/user/project-b\n", nil
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
		{
			Name: "project-b",
			Path: "/home/user/project-b",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-b", Branch: "main"},
			},
		},
	}

	DetectTabs(context.Background(), runner, projects)

	// project-a main should have tab (cwd matches)
	if projects[0].Worktrees[0].TabCount != 1 {
		t.Errorf("project-a/main TabCount = %d, want 1", projects[0].Worktrees[0].TabCount)
	}
	// project-a feature should not have tab
	if projects[0].Worktrees[1].TabCount != 0 {
		t.Errorf("project-a/feature TabCount = %d, want 0", projects[0].Worktrees[1].TabCount)
	}
	// project-b main should have tab
	if projects[1].Worktrees[0].TabCount != 1 {
		t.Errorf("project-b/main TabCount = %d, want 1", projects[1].Worktrees[0].TabCount)
	}
}

func TestDetectTabs_NoGhostty(t *testing.T) {
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  100     1 /bin/zsh
`

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if name == "ps" {
			return psOutput, nil
		}
		return "", fmt.Errorf("should not be called")
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

	DetectTabs(context.Background(), runner, projects)

	if projects[0].Worktrees[0].TabCount != 0 {
		t.Error("should not detect tabs when ghostty is not running")
	}
}

func TestDetectTabs_PSFails(t *testing.T) {
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

	// Should not panic; graceful no-op.
	DetectTabs(context.Background(), runner, projects)

	if projects[0].Worktrees[0].TabCount != 0 {
		t.Error("should not detect tabs when ps fails")
	}
}

func TestDetectTabs_LsofFails(t *testing.T) {
	psOutput := `  PID  PPID COMM
    1     0 /sbin/launchd
  501     1 ghostty
  503   501 -zsh
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

	// Should not panic; graceful skip of that PID.
	DetectTabs(context.Background(), runner, projects)

	if projects[0].Worktrees[0].TabCount != 0 {
		t.Error("should not detect tabs when lsof fails")
	}
}
