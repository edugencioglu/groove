package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme for TUI card components.
// Each theme has Normal (default) and Selected (highlighted) states,
// plus accent colors for status indicators.
//
// Card colors control the worktree card appearance:
//   - Normal state  = default (unfocused) card
//   - Selected state = cursor-highlighted card
//
// Quick reference for manual editing:
//   CardFg / SelectedFg              → worktree name (bold)
//   CardBadgeBg / CardBadgeFg        → [FEATURE] badge background & text
//   CardBranchInfo                   → branch name, ↑n ↓n, git status, separator
//   CardClaudeStatus                 → claude session, disk usage, tab count
//   CardBorder / SelectedBorder      → border around each card
//   AccentPrimary                    → status dot (●) color
type Theme struct {
	Name string

	// Card — Normal state (unfocused)
	CardBg           lipgloss.Color // card background fill
	CardFg           lipgloss.Color // worktree name text (bold)
	CardBorder       lipgloss.Color // border around the card
	CardBadgeBg      lipgloss.Color // [FEATURE], [HOTFIX] badge background
	CardBadgeFg      lipgloss.Color // [FEATURE], [HOTFIX] badge text
	CardBranchInfo   lipgloss.Color // branch name, ↑n ↓n, git status, ─── separator
	CardClaudeStatus lipgloss.Color // claude session, disk usage, tab count

	// Card — Selected state (cursor on this card)
	SelectedBg           lipgloss.Color // card background when selected
	SelectedFg           lipgloss.Color // worktree name when selected (bold)
	SelectedBorder       lipgloss.Color // border when selected
	SelectedBadgeBg      lipgloss.Color // badge background when selected
	SelectedBadgeFg      lipgloss.Color // badge text when selected
	SelectedBranchInfo   lipgloss.Color // branch info when selected
	SelectedClaudeStatus lipgloss.Color // claude status when selected

	// Accents — used for status indicators and badges
	AccentPrimary   lipgloss.Color // status dot (●), primary accent
	AccentSecondary lipgloss.Color // secondary accent (darker shade)
	AccentSuccess   lipgloss.Color // green — clean status
	AccentWarning   lipgloss.Color // yellow — warnings
	AccentDanger    lipgloss.Color // red — errors

	// General UI (shared across all themes)
	Background lipgloss.Color // overall app background
	Foreground lipgloss.Color // default text color
	Subtle     lipgloss.Color // borders, dividers, separators
	Highlight  lipgloss.Color // active elements outside cards
}

// ---------------------------------------------------------------------------
// Flat UI Themes — derived from the Flat UI Color Palette
// ---------------------------------------------------------------------------

var Turquoise = Theme{
	Name: "Turquoise",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#16A086"),
	CardBadgeBg:      lipgloss.Color("#1e1e2e"),
	CardBadgeFg:      lipgloss.Color("#6c7086"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#16A086"),
	SelectedFg:           lipgloss.Color("#1e1e2e"),
	SelectedBorder:       lipgloss.Color("#1ABC9C"),
	SelectedBadgeBg:      lipgloss.Color("#16A086"),
	SelectedBadgeFg:      lipgloss.Color("#2D3E50"),
	SelectedBranchInfo:   lipgloss.Color("#2D3E50"),
	SelectedClaudeStatus: lipgloss.Color("#2D3E50"),

	AccentPrimary:   lipgloss.Color("#1ABC9C"),
	AccentSecondary: lipgloss.Color("#16A086"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#1ABC9C"),
}

var Emerald = Theme{
	Name: "Emerald",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#27AE61"),
	CardBadgeBg:      lipgloss.Color("#2DCC70"),
	CardBadgeFg:      lipgloss.Color("#1e1e2e"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#27AE61"),
	SelectedFg:           lipgloss.Color("#1e1e2e"),
	SelectedBorder:       lipgloss.Color("#2DCC70"),
	SelectedBadgeBg:      lipgloss.Color("#2DCC70"),
	SelectedBadgeFg:      lipgloss.Color("#1e1e2e"),
	SelectedBranchInfo:   lipgloss.Color("#1a3a1a"),
	SelectedClaudeStatus: lipgloss.Color("#1a3a1a"),

	AccentPrimary:   lipgloss.Color("#2DCC70"),
	AccentSecondary: lipgloss.Color("#27AE61"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#2DCC70"),
}

var PeterRiver = Theme{
	Name: "Peter River",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#297FB8"),
	CardBadgeBg:      lipgloss.Color("#3598DB"),
	CardBadgeFg:      lipgloss.Color("#1e1e2e"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#297FB8"),
	SelectedFg:           lipgloss.Color("#ffffff"),
	SelectedBorder:       lipgloss.Color("#3598DB"),
	SelectedBadgeBg:      lipgloss.Color("#3598DB"),
	SelectedBadgeFg:      lipgloss.Color("#1e1e2e"),
	SelectedBranchInfo:   lipgloss.Color("#b8d4e8"),
	SelectedClaudeStatus: lipgloss.Color("#b8d4e8"),

	AccentPrimary:   lipgloss.Color("#3598DB"),
	AccentSecondary: lipgloss.Color("#297FB8"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#3598DB"),
}

var Amethyst = Theme{
	Name: "Amethyst",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#8D44AD"),
	CardBadgeBg:      lipgloss.Color("#1e1e2e"),
	CardBadgeFg:      lipgloss.Color("#6c7086"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#8D44AD"),
	SelectedFg:           lipgloss.Color("#ffffff"),
	SelectedBorder:       lipgloss.Color("#9A59B5"),
	SelectedBadgeBg:      lipgloss.Color("#8D44AD"),
	SelectedBadgeFg:      lipgloss.Color("#d4b8e0"),
	SelectedBranchInfo:   lipgloss.Color("#d4b8e0"),
	SelectedClaudeStatus: lipgloss.Color("#d4b8e0"),

	AccentPrimary:   lipgloss.Color("#9A59B5"),
	AccentSecondary: lipgloss.Color("#8D44AD"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#9A59B5"),
}

var WetAsphalt = Theme{
	Name: "Wet Asphalt",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#2D3E50"),
	CardBadgeBg:      lipgloss.Color("#1e1e2e"),
	CardBadgeFg:      lipgloss.Color("#6c7086"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#34495E"),
	SelectedFg:           lipgloss.Color("#ECF0F1"),
	SelectedBorder:       lipgloss.Color("#5a7a99"),
	SelectedBadgeBg:      lipgloss.Color("#34495E"),
	SelectedBadgeFg:      lipgloss.Color("#95A5A5"),
	SelectedBranchInfo:   lipgloss.Color("#95A5A5"),
	SelectedClaudeStatus: lipgloss.Color("#95A5A5"),

	AccentPrimary:   lipgloss.Color("#34495E"),
	AccentSecondary: lipgloss.Color("#2D3E50"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#34495E"),
}

var SunFlower = Theme{
	Name: "Sun Flower",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#F1C40F"),
	CardBadgeBg:      lipgloss.Color("#1e1e2e"),
	CardBadgeFg:      lipgloss.Color("#6c7086"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#F1C40F"),
	SelectedFg:           lipgloss.Color("#1e1e2e"),
	SelectedBorder:       lipgloss.Color("#F39C11"),
	SelectedBadgeBg:      lipgloss.Color("#F1C40F"),
	SelectedBadgeFg:      lipgloss.Color("#5a4a00"),
	SelectedBranchInfo:   lipgloss.Color("#5a4a00"),
	SelectedClaudeStatus: lipgloss.Color("#5a4a00"),

	AccentPrimary:   lipgloss.Color("#F1C40F"),
	AccentSecondary: lipgloss.Color("#F39C11"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#E67F22"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#F1C40F"),
}

var Carrot = Theme{
	Name: "Carrot",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#D25400"),
	CardBadgeBg:      lipgloss.Color("#E67F22"),
	CardBadgeFg:      lipgloss.Color("#1e1e2e"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#E67F22"),
	SelectedFg:           lipgloss.Color("#1e1e2e"),
	SelectedBorder:       lipgloss.Color("#F39C11"),
	SelectedBadgeBg:      lipgloss.Color("#E67F22"),
	SelectedBadgeFg:      lipgloss.Color("#1e1e2e"),
	SelectedBranchInfo:   lipgloss.Color("#4a2a00"),
	SelectedClaudeStatus: lipgloss.Color("#4a2a00"),

	AccentPrimary:   lipgloss.Color("#E67F22"),
	AccentSecondary: lipgloss.Color("#D25400"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#E67F22"),
}

var Alizarin = Theme{
	Name: "Alizarin",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#C1392B"),
	CardBadgeBg:      lipgloss.Color("#E84C3D"),
	CardBadgeFg:      lipgloss.Color("#1e1e2e"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#C1392B"),
	SelectedFg:           lipgloss.Color("#ffffff"),
	SelectedBorder:       lipgloss.Color("#E84C3D"),
	SelectedBadgeBg:      lipgloss.Color("#E84C3D"),
	SelectedBadgeFg:      lipgloss.Color("#1e1e2e"),
	SelectedBranchInfo:   lipgloss.Color("#ffd5d0"),
	SelectedClaudeStatus: lipgloss.Color("#ffd5d0"),

	AccentPrimary:   lipgloss.Color("#E84C3D"),
	AccentSecondary: lipgloss.Color("#C1392B"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#E84C3D"),
}

var MidnightBlue = Theme{
	Name: "Midnight Blue",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#2D3E50"),
	CardBadgeBg:      lipgloss.Color("#1e1e2e"),
	CardBadgeFg:      lipgloss.Color("#6c7086"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#2D3E50"),
	SelectedFg:           lipgloss.Color("#ECF0F1"),
	SelectedBorder:       lipgloss.Color("#3598DB"),
	SelectedBadgeBg:      lipgloss.Color("#2D3E50"),
	SelectedBadgeFg:      lipgloss.Color("#95A5A5"),
	SelectedBranchInfo:   lipgloss.Color("#95A5A5"),
	SelectedClaudeStatus: lipgloss.Color("#95A5A5"),

	AccentPrimary:   lipgloss.Color("#3598DB"),
	AccentSecondary: lipgloss.Color("#2D3E50"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#3598DB"),
}

var Silver = Theme{
	Name: "Silver",

	CardBg:           lipgloss.Color("#1e1e2e"),
	CardFg:           lipgloss.Color("#cdd6f4"),
	CardBorder:       lipgloss.Color("#7E8C8D"),
	CardBadgeBg:      lipgloss.Color("#1e1e2e"),
	CardBadgeFg:      lipgloss.Color("#6c7086"),
	CardBranchInfo:   lipgloss.Color("#6c7086"),
	CardClaudeStatus: lipgloss.Color("#6c7086"),

	SelectedBg:           lipgloss.Color("#95A5A5"),
	SelectedFg:           lipgloss.Color("#1e1e2e"),
	SelectedBorder:       lipgloss.Color("#BEC3C7"),
	SelectedBadgeBg:      lipgloss.Color("#95A5A5"),
	SelectedBadgeFg:      lipgloss.Color("#2D3E50"),
	SelectedBranchInfo:   lipgloss.Color("#2D3E50"),
	SelectedClaudeStatus: lipgloss.Color("#2D3E50"),

	AccentPrimary:   lipgloss.Color("#BEC3C7"),
	AccentSecondary: lipgloss.Color("#95A5A5"),
	AccentSuccess:   lipgloss.Color("#2DCC70"),
	AccentWarning:   lipgloss.Color("#F1C40F"),
	AccentDanger:    lipgloss.Color("#E84C3D"),

	Background: lipgloss.Color("#11111b"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#313244"),
	Highlight:  lipgloss.Color("#BEC3C7"),
}

// ---------------------------------------------------------------------------
// Theme registry
// ---------------------------------------------------------------------------

var All = map[string]Theme{
	"turquoise":     Turquoise,
	"emerald":       Emerald,
	"peter-river":   PeterRiver,
	"amethyst":      Amethyst,
	"wet-asphalt":   WetAsphalt,
	"sun-flower":    SunFlower,
	"carrot":        Carrot,
	"alizarin":      Alizarin,
	"midnight-blue": MidnightBlue,
	"silver":        Silver,
}

// Names returns all theme names in display order.
func Names() []string {
	return []string{
		"turquoise",
		"emerald",
		"peter-river",
		"amethyst",
		"wet-asphalt",
		"sun-flower",
		"carrot",
		"alizarin",
		"midnight-blue",
		"silver",
	}
}

// Get returns a theme by name. Falls back to Turquoise if not found.
func Get(name string) Theme {
	if t, ok := All[name]; ok {
		return t
	}
	return Turquoise
}
