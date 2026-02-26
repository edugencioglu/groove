package tui

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/edugencioglu/groove/internal/config"
	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/diskusage"
)

func testProjects() []discovery.Project {
	return []discovery.Project{
		{
			Name: "project-a",
			Path: "/home/user/project-a",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-a", Branch: "main", IsClean: true, TabCount: 1, HasClaude: true, DiskUsage: "153M"},
				{Name: "feature", Path: "/home/user/project-a-feature", Branch: "feat/auth", Modified: 3, Untracked: 2, Ahead: 2, DiskUsage: "201M"},
				{Name: "hotfix", Path: "/home/user/project-a-hotfix", Branch: "hotfix/v2.1", Modified: 1, TabCount: 1, DiskUsage: "155M"},
			},
		},
		{
			Name: "project-b",
			Path: "/home/user/project-b",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/home/user/project-b", Branch: "main", IsClean: true, DiskUsage: "148M"},
				{Name: "feature", Path: "/home/user/project-b-feature", Branch: "feat/api", Modified: 1, DiskUsage: "178M"},
			},
		},
	}
}

// noDuModel creates a test model that doesn't fire async disk usage commands.
// Test data already has DiskUsage pre-populated for deterministic golden files.
func noDuModel(projects []discovery.Project) Model {
	m := New(projects)
	// Override duRunner to a noop that never gets called (no paths to fetch
	// since DiskUsage is already populated).
	m.duRunner = diskusage.CmdRunner(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	})
	return m
}

func TestDashboardView(t *testing.T) {
	m := noDuModel(testProjects())
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardNavDown(t *testing.T) {
	m := noDuModel(testProjects())
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardNavTab(t *testing.T) {
	m := noDuModel(testProjects())
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardEmpty(t *testing.T) {
	m := noDuModel(nil)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardNarrow(t *testing.T) {
	m := noDuModel(testProjects())
	// Narrow terminal.
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(50, 24))

	time.Sleep(200 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardManyProjects(t *testing.T) {
	projects := []discovery.Project{
		{Name: "proj-1", Path: "/tmp/p1", Worktrees: []discovery.Worktree{{Name: "main", Path: "/tmp/p1", Branch: "main", IsClean: true, DiskUsage: "10M"}}},
		{Name: "proj-2", Path: "/tmp/p2", Worktrees: []discovery.Worktree{{Name: "main", Path: "/tmp/p2", Branch: "main", IsClean: true, DiskUsage: "20M"}}},
		{Name: "proj-3", Path: "/tmp/p3", Worktrees: []discovery.Worktree{{Name: "main", Path: "/tmp/p3", Branch: "main", IsClean: true, DiskUsage: "30M"}}},
	}
	m := noDuModel(projects)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)
	// Navigate with Tab twice to reach the 3rd project.
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardWithLabels(t *testing.T) {
	labels := config.Labels{
		"/home/user/project-a-feature": "fixing login bug",
		"/home/user/project-b":         "stable release",
	}
	m := noDuModel(testProjects()).WithLabels(labels)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	bts := readFinal(t, tm)
	teatest.RequireEqualOutput(t, bts)
}

func TestDashboardDiskUsage(t *testing.T) {
	// This test verifies the async Cmd/Msg pipeline for disk usage.
	// Disk usage is fetched asynchronously even though it's not displayed in cards.
	projects := []discovery.Project{
		{
			Name: "test-proj",
			Path: "/tmp/test-proj",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/tmp/test-proj", Branch: "main", IsClean: true},
			},
		},
	}

	duRunner := func(_ context.Context, _ string, args ...string) (string, error) {
		path := args[len(args)-1]
		return "42M\t" + path + "\n", nil
	}

	m := New(projects).WithDuRunner(diskusage.CmdRunner(duRunner))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Wait for the card to render (disk usage is fetched but not displayed).
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("main"))
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestDashboardTabCount(t *testing.T) {
	projects := []discovery.Project{
		{
			Name: "myproject",
			Path: "/tmp/myproject",
			Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/tmp/myproject", Branch: "main", IsClean: true, TabCount: 3, DiskUsage: "100M"},
			},
		},
	}
	m := noDuModel(projects)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Wait for "3 tabs" to appear.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("3 tabs"))
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestSplashScreen(t *testing.T) {
	discoverFn := func(ctx context.Context) ([]discovery.Project, error) {
		// Simulate a brief delay.
		time.Sleep(100 * time.Millisecond)
		return []discovery.Project{
			{Name: "test-proj", Path: "/tmp/test", Worktrees: []discovery.Worktree{
				{Name: "main", Path: "/tmp/test", Branch: "main", IsClean: true, DiskUsage: "50M"},
			}},
		}, nil
	}

	m := NewLoading(discoverFn).WithDuRunner(diskusage.CmdRunner(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// First check that we see the splash.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Scanning worktrees"))
	}, teatest.WithDuration(3*time.Second))

	// Then wait for transition to dashboard.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("test-proj"))
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func readFinal(t *testing.T, tm *teatest.TestModel) []byte {
	t.Helper()
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(out); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
