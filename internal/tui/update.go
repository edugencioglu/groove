package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/diskusage"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle loading state messages first.
	if m.loading {
		return m.updateLoading(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tabActionDoneMsg:
		if msg.err != nil {
			m.status = "Error: " + msg.err.Error()
		} else {
			m.status = ""
		}
		return m, nil

	case diskusage.ResultMsg:
		if msg.Err == nil {
			m.applyDiskUsage(msg.Path, msg.Usage)
		}
		return m, nil

	case tickMsg:
		return m, tea.Batch(tickCmd(), m.doRefresh())

	case refreshDoneMsg:
		if msg.err == nil && len(msg.projects) > 0 {
			m.applyRefresh(msg.projects)
		}
		// Kick off disk usage fetch for the (possibly new) worktrees.
		return m, tea.Batch(m.fetchDiskUsage()...)

	case labelSavedMsg:
		if msg.err != nil {
			m.status = "Error saving label: " + msg.err.Error()
		}
		return m, nil

	case tea.KeyMsg:
		// When modal is active, route to modal handler.
		if m.modal != modalNone {
			return m.updateModal(msg)
		}

		// Clear status on any keypress.
		m.status = ""

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			m.moveUp()
		case "down", "j":
			m.moveDown()

		case "tab":
			m.nextTab()
		case "shift+tab":
			m.prevTab()

		case "r":
			return m, tea.Batch(tickCmd(), m.doRefresh())

		case "enter":
			return m, m.doTabAction()

		case "n":
			return m, m.startRename()
		}
	}

	return m, nil
}

// updateLoading handles messages while in the loading/splash state.
func (m Model) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case loadingTickMsg:
		m.loadingDots++
		// Animate progress up to ~85% while discovering.
		if m.pendingProjects == nil && m.progressBar.Percent() < 0.85 {
			cmd := m.progressBar.IncrPercent(0.10)
			return m, tea.Batch(loadingTickCmd(), cmd)
		}
		return m, loadingTickCmd()

	case loadingDoneMsg:
		if msg.err != nil {
			m.loading = false
			m.status = "Error: " + msg.err.Error()
			return m, nil
		}
		// Store projects and animate to 100%, then wait before transitioning.
		m.pendingProjects = msg.projects
		cmd := m.progressBar.SetPercent(1.0)
		return m, tea.Batch(cmd, splashFinishCmd())

	case splashFinishMsg:
		m.loading = false
		m.projects = m.pendingProjects
		m.pendingProjects = nil
		m.lastRefresh = time.Now()
		cmds := []tea.Cmd{tickCmd()}
		cmds = append(cmds, m.fetchDiskUsage()...)
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// updateModal dispatches key events to the active modal.
func (m Model) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case modalRename:
		return m.updateRenaming(msg)
	}
	return m, nil
}

func (m *Model) applyDiskUsage(path, usage string) {
	for i := range m.projects {
		for j := range m.projects[i].Worktrees {
			if m.projects[i].Worktrees[j].Path == path {
				m.projects[i].Worktrees[j].DiskUsage = usage
				return
			}
		}
	}
}

func (m *Model) applyRefresh(projects []discovery.Project) {
	// Preserve disk usage from old data since it arrives async.
	oldDU := make(map[string]string)
	for _, p := range m.projects {
		for _, wt := range p.Worktrees {
			if wt.DiskUsage != "" {
				oldDU[wt.Path] = wt.DiskUsage
			}
		}
	}

	m.projects = projects
	m.lastRefresh = time.Now()
	m.clampRow()

	// Restore cached disk usage for worktrees that still exist.
	for i := range m.projects {
		for j := range m.projects[i].Worktrees {
			path := m.projects[i].Worktrees[j].Path
			if du, ok := oldDU[path]; ok {
				m.projects[i].Worktrees[j].DiskUsage = du
			}
		}
	}
}

func (m Model) doRefresh() tea.Cmd {
	if m.refreshFunc == nil {
		return nil
	}
	fn := m.refreshFunc
	return func() tea.Msg {
		projects, err := fn(context.Background())
		return refreshDoneMsg{projects: projects, err: err}
	}
}

func (m Model) doTabAction() tea.Cmd {
	if m.tabAction == nil || len(m.projects) == 0 {
		return nil
	}

	p := m.projects[m.cursor.col]
	if m.cursor.row >= len(p.Worktrees) {
		return nil
	}
	wt := p.Worktrees[m.cursor.row]

	return func() tea.Msg {
		err := m.tabAction(wt, p.Name)
		return tabActionDoneMsg{err: err}
	}
}

func (m *Model) moveUp() {
	if m.cursor.row > 0 {
		m.cursor.row--
	}
}

func (m *Model) moveDown() {
	if len(m.projects) == 0 {
		return
	}
	maxRow := len(m.projects[m.cursor.col].Worktrees) - 1
	if m.cursor.row < maxRow {
		m.cursor.row++
	}
}

// nextTab switches to the next project tab (wrapping).
func (m *Model) nextTab() {
	if len(m.projects) == 0 {
		return
	}
	m.cursor.col = (m.cursor.col + 1) % len(m.projects)
	m.clampRow()
}

// prevTab switches to the previous project tab (wrapping).
func (m *Model) prevTab() {
	if len(m.projects) == 0 {
		return
	}
	m.cursor.col = (m.cursor.col - 1 + len(m.projects)) % len(m.projects)
	m.clampRow()
}

func (m *Model) clampRow() {
	if len(m.projects) == 0 {
		return
	}
	if m.cursor.col >= len(m.projects) {
		m.cursor.col = len(m.projects) - 1
	}
	maxRow := len(m.projects[m.cursor.col].Worktrees) - 1
	if maxRow < 0 {
		maxRow = 0
	}
	if m.cursor.row > maxRow {
		m.cursor.row = maxRow
	}
}

// selectedWorktree returns the currently selected worktree, or nil if none.
func (m *Model) selectedWorktree() *discovery.Worktree {
	if len(m.projects) == 0 {
		return nil
	}
	p := m.projects[m.cursor.col]
	if m.cursor.row >= len(p.Worktrees) {
		return nil
	}
	return &p.Worktrees[m.cursor.row]
}

func (m *Model) startRename() tea.Cmd {
	wt := m.selectedWorktree()
	if wt == nil {
		return nil
	}

	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 60
	if m.labels != nil {
		ti.SetValue(m.labels[wt.Path])
	}
	ti.Focus()

	m.modal = modalRename
	m.renameInput = ti
	return textinput.Blink
}

func (m Model) updateRenaming(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		wt := m.selectedWorktree()
		if wt != nil {
			if m.labels == nil {
				m.labels = make(map[string]string)
			}
			val := m.renameInput.Value()
			if val == "" {
				delete(m.labels, wt.Path)
			} else {
				m.labels[wt.Path] = val
			}
			m.modal = modalNone
			return m, m.doSaveLabels()
		}
		m.modal = modalNone
		return m, nil

	case tea.KeyEsc:
		m.modal = modalNone
		return m, nil
	}

	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}

func (m Model) doSaveLabels() tea.Cmd {
	if m.labelSaveFn == nil {
		return nil
	}
	fn := m.labelSaveFn
	labels := make(map[string]string, len(m.labels))
	for k, v := range m.labels {
		labels[k] = v
	}
	return func() tea.Msg {
		return labelSavedMsg{err: fn(labels)}
	}
}
