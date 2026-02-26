package diskusage

import (
	"context"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// ParseDuOutput tests
// ---------------------------------------------------------------------------

func TestParseDuOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "153M\t/home/user/project\n", "153M"},
		{"with spaces", "  4.2G\t/path/to/dir  \n", "4.2G"},
		{"kilobytes", "512K\t/small/dir\n", "512K"},
		{"empty", "", ""},
		{"whitespace only", "  \n  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDuOutput(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchOne tests (mock runner)
// ---------------------------------------------------------------------------

func TestFetchOne(t *testing.T) {
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		return "201M\t/home/user/project\n", nil
	}

	cmd := FetchOne(context.Background(), runner, "/home/user/project")
	msg := cmd()

	result, ok := msg.(ResultMsg)
	if !ok {
		t.Fatalf("expected ResultMsg, got %T", msg)
	}
	if result.Path != "/home/user/project" {
		t.Errorf("path = %q, want /home/user/project", result.Path)
	}
	if result.Usage != "201M" {
		t.Errorf("usage = %q, want 201M", result.Usage)
	}
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
}

func TestFetchOne_Error(t *testing.T) {
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		return "", fmt.Errorf("du failed")
	}

	cmd := FetchOne(context.Background(), runner, "/bad/path")
	msg := cmd()

	result, ok := msg.(ResultMsg)
	if !ok {
		t.Fatalf("expected ResultMsg, got %T", msg)
	}
	if result.Err == nil {
		t.Error("expected error")
	}
}

// ---------------------------------------------------------------------------
// FetchAll tests
// ---------------------------------------------------------------------------

func TestFetchAll(t *testing.T) {
	callCount := 0
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		callCount++
		return "100M\t" + args[len(args)-1] + "\n", nil
	}

	paths := []string{"/a", "/b", "/c"}
	cmds := FetchAll(context.Background(), runner, paths)

	if len(cmds) != 3 {
		t.Fatalf("got %d cmds, want 3", len(cmds))
	}

	// Execute each cmd and verify.
	for i, cmd := range cmds {
		msg := cmd().(ResultMsg)
		if msg.Path != paths[i] {
			t.Errorf("cmd[%d].Path = %q, want %q", i, msg.Path, paths[i])
		}
		if msg.Usage != "100M" {
			t.Errorf("cmd[%d].Usage = %q, want 100M", i, msg.Usage)
		}
	}
}

// ---------------------------------------------------------------------------
// Verify ResultMsg implements tea.Msg
// ---------------------------------------------------------------------------

func TestResultMsg_IsTeaMsg(t *testing.T) {
	var _ tea.Msg = ResultMsg{}
}
