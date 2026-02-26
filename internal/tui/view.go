package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/edugencioglu/groove/internal/discovery"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.loading {
		return m.viewSplash()
	}

	if len(m.projects) == 0 {
		return m.renderHeader() + "\n\n  No projects found.\n"
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteByte('\n')
	b.WriteByte('\n')

	// Tab bar + window content
	b.WriteString(m.renderTabWindow())
	b.WriteByte('\n')

	dashboard := b.String()

	// Modal overlay
	if m.modal == modalRename {
		return m.renderRenameModal(dashboard)
	}

	return dashboard
}

// renderTabWindow renders the tab bar and a connected window border containing
// worktree cards and footer, following the canonical lipgloss tabs pattern.
func (m Model) renderTabWindow() string {
	w := m.width
	if w <= 0 {
		w = 80
	}

	tabRow := m.renderTabBar()
	tabRowWidth := lipgloss.Width(tabRow)

	// Build window content: cards + status + footer separator + footer
	var content strings.Builder
	p := m.projects[m.cursor.col]
	cardInner := m.cardWidth()
	for i, wt := range p.Worktrees {
		selected := m.cursor.row == i
		card := m.renderCard(wt, selected, cardInner, i)
		content.WriteString(card)
		content.WriteByte('\n')
	}

	// Status message
	if m.status != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(colorError).Render(m.status))
		content.WriteByte('\n')
	}

	// Footer separator + help text inside window
	innerWidth := tabRowWidth - 4 // 2 border + 2 padding
	if innerWidth < 20 {
		innerWidth = 20
	}
	content.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("─", innerWidth)))
	content.WriteByte('\n')
	content.WriteString(footerStyle.Render("[tab/shift+tab] switch  [↑↓] navigate  [enter] open  [n] rename  [r] refresh  [q] quit"))

	// Window style: border with no top (connects to tab row above)
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop().
		BorderForeground(colorAccent).
		Width(tabRowWidth - 2). // -2 for left+right border chars
		Padding(0, 1)

	window := windowStyle.Render(content.String())

	return tabRow + "\n" + window
}

// renderHeader returns the multi-line header block with tree icon.
func (m Model) renderHeader() string {
	w := m.width
	if w <= 0 {
		w = 80
	}

	sep := "  "

	// Tree icon lines, each padded to 5 visible chars for alignment.
	tree1 := leafStyle.Render(" ▄█▄") + " "
	tree2 := leafStyle.Render("█████")
	tree3 := stemStyle.Render("  █") + "  "

	// Line 1: tree-top + title + refresh
	titleText := titleAccentStyle.Render("groove") + headerSubtitleStyle.Render(" — Worktree Dashboard")
	line1 := " " + tree1 + sep + titleText
	ago := m.refreshAgo()
	if ago != "" {
		refresh := dimStyle.Render("↻ " + ago)
		gap := w - lipgloss.Width(line1) - lipgloss.Width(refresh) - 1
		if gap < 1 {
			gap = 1
		}
		line1 += strings.Repeat(" ", gap) + refresh
	}

	// Line 2: tree-middle + stats
	stats := dimStyle.Render(fmt.Sprintf("%d projects · %d worktrees", len(m.projects), m.totalWorktrees()))
	line2 := " " + tree2 + sep + stats

	// Line 3: tree-stem
	line3 := " " + tree3

	return line1 + "\n" + line2 + "\n" + line3
}

// totalWorktrees counts all worktrees across projects.
func (m Model) totalWorktrees() int {
	n := 0
	for _, p := range m.projects {
		n += len(p.Worktrees)
	}
	return n
}

// renderTabBar renders the horizontal project tab bar using the canonical
// lipgloss tab pattern: active tab has open bottom, inactive tabs have
// closed bottom borders that form a continuous line.
func (m Model) renderTabBar() string {
	if len(m.projects) == 0 {
		return ""
	}

	w := m.width
	if w <= 0 {
		w = 80
	}

	// Render tabs with first-tab corner overrides so the tab bar bottom
	// merges into the window border below (├/│ connect to window left │).
	var renderedTabs []string
	for i, p := range m.projects {
		isActive := i == m.cursor.col
		isFirst := i == 0

		style := tabInactiveStyle
		if isActive {
			style = tabActiveStyle
		}

		if isFirst {
			border, _, _, _, _ := style.GetBorder()
			if isActive {
				border.BottomLeft = "│"
			} else {
				border.BottomLeft = "├"
			}
			style = style.Border(border)
		}

		renderedTabs = append(renderedTabs, style.Render(p.Name))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	rowWidth := lipgloss.Width(row)
	gapAvailable := w - rowWidth

	if gapAvailable > 0 {
		// Gap filler: two-line string aligned at bottom with tabs.
		// Line 1 (content height): invisible spaces.
		// Line 2 (bottom):         ─── ending with ┐ to connect to window right │.
		gapDashes := gapAvailable - 1
		if gapDashes < 0 {
			gapDashes = 0
		}
		topLine := strings.Repeat(" ", gapAvailable)
		bottomLine := tabGapStyle.Render(strings.Repeat("─", gapDashes) + "┐")
		gap := topLine + "\n" + bottomLine
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	} else {
		// No gap — tabs fill the width. Override last tab's right corner
		// to connect to the window right │ border.
		lastIdx := len(m.projects) - 1
		isActive := lastIdx == m.cursor.col
		style := tabInactiveStyle
		if isActive {
			style = tabActiveStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isActive {
			border.BottomRight = "│"
		} else {
			border.BottomRight = "┤"
		}
		if lastIdx == 0 { // single tab: also first
			if isActive {
				border.BottomLeft = "│"
			} else {
				border.BottomLeft = "├"
			}
		}
		style = style.Border(border)
		renderedTabs[lastIdx] = style.Render(m.projects[lastIdx].Name)
		row = lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	}

	return row
}

// cardWidth calculates the inner width for worktree cards.
// Accounts for: window border (2) + window padding (2) + card border (2) + card padding (2) = 8
func (m Model) cardWidth() int {
	w := m.width
	if w <= 0 {
		w = 80
	}
	inner := w - 8
	if inner < 20 {
		inner = 20
	}
	return inner
}

func (m Model) renderCard(wt discovery.Worktree, selected bool, innerWidth int, _ int) string {
	ci := cardInfoForWorktree(wt)
	th := ci.theme

	// Pick colors based on selected state.
	var bgColor, fgColor, borderColor lipgloss.Color
	var badgeBgColor, badgeFgColor lipgloss.Color
	var branchInfoColor, claudeStatusColor lipgloss.Color
	if selected {
		bgColor = th.SelectedBg
		fgColor = th.SelectedFg
		borderColor = th.SelectedBorder
		badgeBgColor = th.SelectedBadgeBg
		badgeFgColor = th.SelectedBadgeFg
		branchInfoColor = th.SelectedBranchInfo
		claudeStatusColor = th.SelectedClaudeStatus
	} else {
		bgColor = th.CardBg
		fgColor = th.CardFg
		borderColor = th.CardBorder
		badgeBgColor = th.CardBadgeBg
		badgeFgColor = th.CardBadgeFg
		branchInfoColor = th.CardBranchInfo
		claudeStatusColor = th.CardClaudeStatus
	}

	// All inner styles carry Background so ANSI resets don't clear the card bg.
	nameStyle := lipgloss.NewStyle().Foreground(fgColor).Background(bgColor).Bold(true)
	badgeStyle := lipgloss.NewStyle().Foreground(badgeFgColor).Background(badgeBgColor)
	branchInfoStyle := lipgloss.NewStyle().Foreground(branchInfoColor).Background(bgColor).Faint(true)
	claudeStatusStyle := lipgloss.NewStyle().Foreground(claudeStatusColor).Background(bgColor).Faint(true)
	accentStyle := lipgloss.NewStyle().Foreground(th.AccentPrimary).Background(bgColor)
	bgStyle := lipgloss.NewStyle().Background(bgColor)

	// Text area = innerWidth minus card padding (1 left + 1 right).
	textWidth := innerWidth - 2
	if textWidth < 10 {
		textWidth = 10
	}

	// bgPad returns n spaces with the card background color.
	bgPad := func(n int) string {
		if n <= 0 {
			return ""
		}
		return bgStyle.Render(strings.Repeat(" ", n))
	}

	var lines []string

	// Resolve display name: custom label or worktree name.
	cardName := wt.Name
	if m.labels != nil {
		if label, ok := m.labels[wt.Path]; ok && label != "" {
			cardName = label
		}
	}

	// Line 1: name [CATEGORY]                              [SEL]
	leftParts := bgPad(1) + nameStyle.Render(cardName)
	if ci.category != "" {
		badge := badgeStyle.Render("[" + strings.ToUpper(ci.category) + "]")
		leftParts += bgPad(1) + badge
	}
	if selected {
		selBadge := badgeStyle.Render("[SEL]")
		lines = append(lines, bgPadLineWithRight(leftParts, selBadge+bgPad(1), innerWidth, bgPad))
	} else {
		lines = append(lines, bgPadLine(leftParts, innerWidth, bgPad))
	}

	// Line 2: ⎇ branch-name ↑n ↓n
	branchText := "⎇ " + wt.Branch
	if wt.Ahead > 0 || wt.Behind > 0 {
		branchText += fmt.Sprintf(" ↑%d ↓%d", wt.Ahead, wt.Behind)
	}
	lines = append(lines, bgPadLine(bgPad(1)+branchInfoStyle.Render(branchText), innerWidth, bgPad))

	// Line 3: ● git status
	var statusText string
	if wt.IsClean {
		statusText = accentStyle.Render("●") + bgPad(1) + branchInfoStyle.Render("clean")
	} else {
		var parts []string
		if wt.Modified > 0 {
			parts = append(parts, fmt.Sprintf("%d modified", wt.Modified))
		}
		if wt.Untracked > 0 {
			parts = append(parts, fmt.Sprintf("%d untracked", wt.Untracked))
		}
		if len(parts) == 0 {
			parts = append(parts, "dirty")
		}
		statusText = accentStyle.Render("●") + bgPad(1) + branchInfoStyle.Render(strings.Join(parts, ", "))
	}
	lines = append(lines, bgPadLine(bgPad(1)+statusText, innerWidth, bgPad))

	// Line 4: separator
	sepWidth := textWidth - 2
	if sepWidth < 1 {
		sepWidth = 1
	}
	lines = append(lines, bgPadLine(bgPad(1)+branchInfoStyle.Render(strings.Repeat("─", sepWidth)), innerWidth, bgPad))

	// Line 5: status bar — tabs • claude • disk
	var indicators []string
	if wt.TabCount > 0 {
		label := "1 tab"
		if wt.TabCount > 1 {
			label = fmt.Sprintf("%d tabs", wt.TabCount)
		}
		indicators = append(indicators, label)
	}
	if wt.HasClaude {
		indicators = append(indicators, "● claude:active")
	} else {
		indicators = append(indicators, "○ no session")
	}
	if wt.DiskUsage != "" {
		indicators = append(indicators, formatDiskUsage(wt.DiskUsage))
	}
	statusBar := claudeStatusStyle.Render(strings.Join(indicators, " • "))
	lines = append(lines, bgPadLine(bgPad(1)+statusBar, innerWidth, bgPad))

	content := strings.Join(lines, "\n")

	style := cardBaseStyle.BorderForeground(borderColor)

	return style.Width(innerWidth).Render(content)
}

// bgPadLine pads a rendered string to exactly width visible characters
// using background-colored spaces from bgPad.
func bgPadLine(s string, width int, bgPad func(int) string) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + bgPad(width-w)
}

// bgPadLineWithRight places left and right content on the same line,
// filling the gap with background-colored spaces.
func bgPadLineWithRight(left, right string, width int, bgPad func(int) string) string {
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := width - lw - rw
	if gap < 1 {
		gap = 1
	}
	return left + bgPad(gap) + right
}

// formatDiskUsage inserts a space before the unit suffix for readability.
// "841M" → "841 MB", "1.2G" → "1.2 GB", etc.
func formatDiskUsage(s string) string {
	if len(s) < 2 {
		return s
	}
	unit := s[len(s)-1]
	switch unit {
	case 'K', 'M', 'G', 'T':
		return s[:len(s)-1] + " " + string(unit) + "B"
	default:
		return s
	}
}


func (m Model) refreshAgo() string {
	if m.lastRefresh.IsZero() {
		return ""
	}
	d := time.Since(m.lastRefresh).Truncate(time.Second)
	if d < time.Second {
		return "just now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm ago", int(d.Minutes()))
}

// renderRenameModal overlays a centered rename modal on top of the dashboard
// using PlaceOverlay so the background content remains visible around it.
func (m Model) renderRenameModal(dashboard string) string {
	wt := m.selectedWorktree()
	if wt == nil {
		return dashboard
	}

	title := modalTitleStyle.Render("Rename: " + wt.Name)
	input := m.renameInput.View()
	hint := modalHintStyle.Render("[enter] save  [esc] cancel")

	content := title + "\n\n" + input + "\n\n" + hint
	box := modalStyle.Render(content)

	w := m.width
	h := m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	// Pad the dashboard to fill the full terminal height so the overlay
	// has a complete background to composite onto.
	bg := padToHeight(dashboard, h)

	// Center the modal over the background.
	fgW := lipgloss.Width(box)
	fgH := lipgloss.Height(box)
	x := (w - fgW) / 2
	y := (h - fgH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return placeOverlay(x, y, box, bg)
}

// viewSplash renders the loading/splash screen.
func (m Model) viewSplash() string {
	w := m.width
	h := m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	// Pad each tree line to 5 chars so lipgloss.Place keeps them aligned.
	treeLine1 := leafStyle.Render(" ▄█▄") + " "
	treeLine2 := leafStyle.Render("█████")
	treeLine3 := " " + stemStyle.Render(" █") + "  "
	tree := treeLine1 + "\n" + treeLine2 + "\n" + treeLine3
	title := splashTitleStyle.Render("groove")
	subtitle := splashSubtitleStyle.Render("Worktree Dashboard")

	dots := strings.Repeat(".", (m.loadingDots%3)+1)
	loading := dimStyle.Render("Scanning worktrees" + dots)

	bar := m.progressBar.View()

	content := tree + "\n" + title + "\n" + subtitle + "\n\n" + loading + "\n\n" + bar

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// placeOverlay composites fg on top of bg at position (x, y).
// It is ANSI-aware: background content is preserved on both sides of the
// overlay using charmbracelet/x/ansi truncation functions.
func placeOverlay(x, y int, fg, bg string) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	for i, fgLine := range fgLines {
		bgIdx := y + i
		if bgIdx < 0 || bgIdx >= len(bgLines) {
			continue
		}

		bgLine := bgLines[bgIdx]
		fgWidth := ansi.StringWidth(fgLine)
		bgWidth := ansi.StringWidth(bgLine)

		var sb strings.Builder

		// Left part of background (first x visible chars).
		if x > 0 {
			left := ansi.Truncate(bgLine, x, "")
			sb.WriteString(left)
			leftW := ansi.StringWidth(left)
			if leftW < x {
				sb.WriteString(strings.Repeat(" ", x-leftW))
			}
		}

		// Foreground content.
		sb.WriteString(fgLine)

		// Right part of background (everything after overlay region).
		rightStart := x + fgWidth
		if rightStart < bgWidth {
			right := ansi.TruncateLeft(bgLine, rightStart, "")
			sb.WriteString(right)
		}

		bgLines[bgIdx] = sb.String()
	}

	return strings.Join(bgLines, "\n")
}

// padToHeight ensures the string has at least h lines.
func padToHeight(s string, h int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
