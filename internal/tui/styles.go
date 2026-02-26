package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/theme"
)

// Tab border definitions for the canonical lipgloss tab pattern.
// Active tab has open bottom (spaces), inactive has closed bottom.
var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	inactiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}
)

// ---------------------------------------------------------------------------
// UI Color Palette — edit these to restyle the entire UI
// ---------------------------------------------------------------------------
//
// These colors control everything outside of card theming (cards use
// internal/theme). Change a color here and every style that uses it updates.
var (
	colorAccent    = lipgloss.Color("#2ECC71") // green — borders, tab highlights, title, modal frames
	colorMuted     = lipgloss.Color("#6b7280") // gray   — secondary text, footer, hints, subtitles
	colorBrightTxt = lipgloss.Color("#e5e7eb") // white  — active tab text, modal title, header text
	colorError     = lipgloss.Color("#E84C3D") // red    — status/error messages
)

// ---------------------------------------------------------------------------
// Logo — tree icon with green leaves and brown stem
// ---------------------------------------------------------------------------

var (
	colorLeaf = lipgloss.Color("#2ECC71") // green — tree leaves
	colorStem = lipgloss.Color("#8B6914") // brown — tree stem

	leafStyle = lipgloss.NewStyle().Foreground(colorLeaf)
	stemStyle = lipgloss.NewStyle().Foreground(colorStem)
)

// ---------------------------------------------------------------------------
// Styles — grouped by UI region
// ---------------------------------------------------------------------------
var (
	// Header
	titleAccentStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	headerSubtitleStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Footer
	footerStyle = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1)

	// Tab bar
	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBrightTxt).
			Border(activeTabBorder, true).
			BorderForeground(colorAccent).
			Padding(0, 1)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Border(inactiveTabBorder, true).
				BorderForeground(colorAccent).
				Padding(0, 1)

	tabGapStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	// Cards
	cardBaseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			MarginBottom(1)

	// General
	dimStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Modal
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBrightTxt)
	modalHintStyle  = lipgloss.NewStyle().Foreground(colorMuted)

	// Splash screen
	splashTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	splashSubtitleStyle = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
)

// cardInfo holds the resolved theme and category label for a worktree card.
type cardInfo struct {
	theme    theme.Theme
	category string // e.g., "feature", "hotfix", "ai-issues", "clean"
}

// cardInfoForWorktree picks the theme and category based on the worktree name.
func cardInfoForWorktree(wt discovery.Worktree) cardInfo {
	name := strings.ToLower(wt.Name)

	switch {
	case strings.Contains(name, "hotfix"):
		return cardInfo{theme: theme.Alizarin, category: "hotfix"}
	case strings.Contains(name, "feature"):
		return cardInfo{theme: theme.PeterRiver, category: "feature"}
	case strings.Contains(name, "ai-") || strings.Contains(name, "issue"):
		return cardInfo{theme: theme.Carrot, category: "ai-issues"}
	case wt.IsClean:
		return cardInfo{theme: theme.Emerald, category: "clean"}
	default:
		return cardInfo{theme: theme.Emerald, category: ""}
	}
}
