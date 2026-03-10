package tabs

import (
	"strings"
	"testing"
)

func TestNewTabScript(t *testing.T) {
	script := NewTabScript("/Users/emre/project-a")

	if !strings.Contains(script, `tell application "Ghostty"`) {
		t.Error("script should target Ghostty directly")
	}
	if !strings.Contains(script, `activate`) {
		t.Error("script should activate Ghostty")
	}
	if !strings.Contains(script, `new surface configuration`) {
		t.Error("script should create surface configuration")
	}
	if !strings.Contains(script, `initial working directory of cfg to "/Users/emre/project-a"`) {
		t.Error("script should set working directory in config")
	}
	if !strings.Contains(script, `new tab in front window with configuration cfg`) {
		t.Error("script should create new tab with config")
	}
	// Should NOT use System Events anymore.
	if strings.Contains(script, `System Events`) {
		t.Error("script should not use System Events (use native Ghostty API)")
	}
}

func TestNewTabScript_PathWithSpaces(t *testing.T) {
	script := NewTabScript("/Users/emre/my project/src")

	if !strings.Contains(script, `initial working directory of cfg to "/Users/emre/my project/src"`) {
		t.Error("script should preserve spaces in path (handled by AppleScript string)")
	}
}

func TestNewTabScript_PathWithQuotes(t *testing.T) {
	script := NewTabScript(`/Users/emre/it's "special"`)

	if !strings.Contains(script, `it's \"special\"`) {
		t.Error("script should escape double quotes in path")
	}
}

func TestJumpTabScript(t *testing.T) {
	script := JumpTabScript("/Users/emre/projects/project-a/feature")

	if !strings.Contains(script, `tell application "Ghostty"`) {
		t.Error("script should target Ghostty directly")
	}
	if !strings.Contains(script, `activate`) {
		t.Error("script should activate Ghostty")
	}
	if !strings.Contains(script, `working directory of term is "/Users/emre/projects/project-a/feature"`) {
		t.Error("script should match terminal working directory")
	}
	if !strings.Contains(script, `focus term`) {
		t.Error("script should focus the matching terminal")
	}
	if !strings.Contains(script, `error "no matching tab found"`) {
		t.Error("script should error when no tab matches")
	}
	// Should NOT use System Events anymore.
	if strings.Contains(script, `System Events`) {
		t.Error("script should not use System Events (use native Ghostty API)")
	}
}

func TestJumpTabScript_PathWithQuotes(t *testing.T) {
	script := JumpTabScript(`/Users/emre/it's "special"`)

	if !strings.Contains(script, `working directory of term is "/Users/emre/it's \"special\""`) {
		t.Error("script should escape quotes in path")
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
