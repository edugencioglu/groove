package tabs

import (
	"strings"
	"testing"
)

func TestNewTabScript(t *testing.T) {
	script := NewTabScript("/Users/emre/project-a")

	if !strings.Contains(script, `tell application "Ghostty" to activate`) {
		t.Error("script should activate Ghostty")
	}
	if !strings.Contains(script, `keystroke "t" using command down`) {
		t.Error("script should open new tab with Cmd+T")
	}
	if !strings.Contains(script, `cd /Users/emre/project-a && clear`) {
		t.Error("script should cd to the worktree path")
	}
	if !strings.Contains(script, `key code 36`) {
		t.Error("script should press Enter")
	}
	if !strings.Contains(script, `delay 0.3`) {
		t.Error("script should include delay for tab to open")
	}
}

func TestNewTabScript_PathWithSpaces(t *testing.T) {
	script := NewTabScript("/Users/emre/my project/src")

	if !strings.Contains(script, `cd /Users/emre/my project/src && clear`) {
		t.Error("script should preserve spaces in path")
	}
}

func TestNewTabScript_PathWithQuotes(t *testing.T) {
	script := NewTabScript(`/Users/emre/it's "special"`)

	if !strings.Contains(script, `it's \"special\"`) {
		t.Error("script should escape double quotes in path")
	}
}

func TestJumpTabScript(t *testing.T) {
	script := JumpTabScript("project-a", "feature", "")

	if !strings.Contains(script, `tell application "Ghostty" to activate`) {
		t.Error("script should activate Ghostty")
	}
	// Uses `name of` instead of `title of` to avoid Unicode errors with ✳.
	if !strings.Contains(script, `name of t`) {
		t.Error("script should use name (not title) to avoid Unicode errors")
	}
	if !strings.Contains(script, `tabName contains "project-a"`) {
		t.Error("script should check for project name")
	}
	if !strings.Contains(script, `tabName contains "feature"`) {
		t.Error("script should check for worktree name")
	}
	if !strings.Contains(script, `click t`) {
		t.Error("script should click matching tab")
	}
	if !strings.Contains(script, `perform action "AXRaise"`) {
		t.Error("script should raise the matching window")
	}
	if !strings.Contains(script, `error "no matching tab found"`) {
		t.Error("script should error when no tab matches")
	}
}

func TestJumpTabScript_HeaderFallback(t *testing.T) {
	script := JumpTabScript("project-a", "feature", "Documents/projects/project-a/feature")

	// Pass 2: header-based matching reads first 1000 chars of pane content.
	if !strings.Contains(script, `text 1 thru 1000 of taVal`) {
		t.Error("script should limit content check to first 1000 chars")
	}
	if !strings.Contains(script, `taVal contains "Documents/projects/project-a/feature"`) {
		t.Error("script should check pane header for relative worktree path")
	}
	if !strings.Contains(script, `group 1 of group 1 of front window`) {
		t.Error("script should use fresh front window ref for content area")
	}
}

func TestJumpTabScript_LabelWithQuotes(t *testing.T) {
	script := JumpTabScript("proj", `"special"`, "")

	if !strings.Contains(script, `tabName contains "\"special\""`) {
		t.Error("script should escape quotes in label")
	}
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`hello`, `hello`},
		{`say "hi"`, `say \"hi\"`},
		{`path\to\file`, `path\\to\\file`},
		{`"quotes" and \backslash`, `\"quotes\" and \\backslash`},
	}

	for _, tt := range tests {
		got := escapeAppleScript(tt.input)
		if got != tt.want {
			t.Errorf("escapeAppleScript(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
