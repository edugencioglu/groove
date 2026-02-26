package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edugencioglu/groove/internal/config"
	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/diskusage"
)

// TabAction is a function that executes a tab action (create or jump).
// It receives the worktree and project name for context.
type TabAction func(wt discovery.Worktree, projectName string) error

// RefreshFunc is called on auto-refresh to reload projects with all enrichment.
// It returns the refreshed projects.
type RefreshFunc func(ctx context.Context) ([]discovery.Project, error)

// LabelSaveFn persists labels to disk.
type LabelSaveFn func(config.Labels) error

// DiscoverFunc performs initial project discovery (used by splash screen).
type DiscoverFunc func(ctx context.Context) ([]discovery.Project, error)

// labelSavedMsg is sent when a label save completes.
type labelSavedMsg struct{ err error }

// modalKind represents the active modal overlay.
type modalKind int

const (
	modalNone   modalKind = iota
	modalRename           // rename worktree label
)

// Model is the Bubbletea model for the worktree dashboard.
type Model struct {
	projects    []discovery.Project
	cursor      cursor
	width       int
	height      int
	tabAction   TabAction
	refreshFunc RefreshFunc
	status      string    // status message shown briefly after actions
	lastRefresh time.Time // when data was last loaded
	duRunner    diskusage.CmdRunner

	labels      config.Labels
	modal       modalKind
	renameInput textinput.Model
	labelSaveFn LabelSaveFn

	// Splash/loading state
	loading         bool
	loadingDots     int // 0-2 for dot animation
	progressBar     progress.Model
	discoverFn      DiscoverFunc
	pendingProjects []discovery.Project // held until splash finishes
}

type cursor struct {
	col int // project index (tab bar)
	row int // worktree index within project
}

const refreshInterval = 30 * time.Second

// tabActionDoneMsg is sent when a tab action completes.
type tabActionDoneMsg struct {
	err error
}

// tickMsg triggers auto-refresh.
type tickMsg time.Time

// refreshDoneMsg carries refreshed project data.
type refreshDoneMsg struct {
	projects []discovery.Project
	err      error
}

// loadingTickMsg triggers loading animation update.
type loadingTickMsg struct{}

// loadingDoneMsg signals that initial discovery is complete.
type loadingDoneMsg struct {
	projects []discovery.Project
	err      error
}

// splashFinishMsg is sent after progress reaches 100% and a short delay.
type splashFinishMsg struct{}

// New creates a new TUI model with the given projects.
func New(projects []discovery.Project) Model {
	return Model{
		projects:    projects,
		lastRefresh: time.Now(),
		duRunner:    diskusage.DefaultCmdRunner,
	}
}

// NewLoading creates a model that starts in the loading/splash state.
// It will call discoverFn asynchronously in Init().
func NewLoading(discoverFn DiscoverFunc) Model {
	p := progress.New(progress.WithScaledGradient("#2ECC71", "#58D68D"))
	p.Width = 30
	return Model{
		loading:     true,
		discoverFn:  discoverFn,
		duRunner:    diskusage.DefaultCmdRunner,
		progressBar: p,
	}
}

// WithTabAction sets the function called when Enter is pressed.
func (m Model) WithTabAction(fn TabAction) Model {
	m.tabAction = fn
	return m
}

// WithRefreshFunc sets the function called on auto-refresh.
func (m Model) WithRefreshFunc(fn RefreshFunc) Model {
	m.refreshFunc = fn
	return m
}

// WithDuRunner sets the command runner for disk usage (for testing).
func (m Model) WithDuRunner(runner diskusage.CmdRunner) Model {
	m.duRunner = runner
	return m
}

// WithLabels sets the initial label map.
func (m Model) WithLabels(labels config.Labels) Model {
	m.labels = labels
	return m
}

// WithLabelSaveFn sets the function used to persist labels.
func (m Model) WithLabelSaveFn(fn LabelSaveFn) Model {
	m.labelSaveFn = fn
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.loading {
		return tea.Batch(m.doDiscover(), loadingTickCmd())
	}

	cmds := []tea.Cmd{tickCmd()}

	// Kick off async disk usage fetch for all worktrees.
	cmds = append(cmds, m.fetchDiskUsage()...)

	return tea.Batch(cmds...)
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadingTickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return loadingTickMsg{}
	})
}

func (m Model) doDiscover() tea.Cmd {
	if m.discoverFn == nil {
		return nil
	}
	fn := m.discoverFn
	return func() tea.Msg {
		projects, err := fn(context.Background())
		return loadingDoneMsg{projects: projects, err: err}
	}
}

func splashFinishCmd() tea.Cmd {
	return tea.Tick(700*time.Millisecond, func(time.Time) tea.Msg {
		return splashFinishMsg{}
	})
}

// worktreePathsMissingDU returns paths of worktrees that don't have disk usage yet.
func (m Model) worktreePathsMissingDU() []string {
	var paths []string
	for _, p := range m.projects {
		for _, wt := range p.Worktrees {
			if wt.DiskUsage == "" {
				paths = append(paths, wt.Path)
			}
		}
	}
	return paths
}

// fetchDiskUsage returns Cmds to fetch disk usage for worktrees missing it.
func (m Model) fetchDiskUsage() []tea.Cmd {
	paths := m.worktreePathsMissingDU()
	if len(paths) == 0 {
		return nil
	}
	return diskusage.FetchAll(context.Background(), m.duRunner, paths)
}
