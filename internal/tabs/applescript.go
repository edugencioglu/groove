package tabs

import (
	"fmt"
	"strings"
)

// NewTabScript generates an AppleScript that opens a new Ghostty tab and
// changes directory to the given path.
func NewTabScript(path string) string {
	escaped := escapeAppleScript(path)
	return fmt.Sprintf(`tell application "Ghostty" to activate
tell application "System Events"
    tell process "Ghostty"
        keystroke "t" using command down
        delay 0.3
        keystroke "cd %s && clear"
        key code 36
    end tell
end tell`, escaped)
}

// JumpTabScript generates an AppleScript that switches to an existing Ghostty
// tab for the given worktree. It uses a two-pass strategy:
//
//  1. Title match: search all windows for a tab whose name contains both the
//     project name and worktree name (fast, no side effects).
//  2. Header match: click each unmatched tab and read the first 1000 characters
//     of each split pane's content. Claude Code always shows the working
//     directory in its header (e.g. "~/Documents/.../guido/ai-issues"), so
//     checking the top of the pane reliably identifies the worktree without
//     false positives from scrollback history.
//
// The relPath parameter should be the worktree path relative to $HOME
// (e.g. "Documents/insider-projects/guido/ai-issues") so it matches both
// "~/" and absolute path formats.
//
// Uses `name of` instead of `title of` because AppleScript's title accessor
// fails on the ✳ character that Claude Code prepends to tab titles.
//
// Returns an AppleScript error if no matching tab is found.
func JumpTabScript(project, worktree, relPath string) string {
	escProject := escapeAppleScript(project)
	escWorktree := escapeAppleScript(worktree)
	escRelPath := escapeAppleScript(relPath)
	return fmt.Sprintf(`tell application "Ghostty" to activate
tell application "System Events"
    -- Pass 1: name-based matching (fast, no side effects).
    tell process "Ghostty"
        set wCount to count of windows
        repeat with wIdx from 1 to wCount
            try
                set tabGroup to tab group 1 of window wIdx
                set allTabs to radio buttons of tabGroup
                repeat with t in allTabs
                    set tabName to name of t
                    if tabName contains "%s" and tabName contains "%s" then
                        click t
                        perform action "AXRaise" of window wIdx
                        return
                    end if
                end repeat
            end try
        end repeat
    end tell

    -- Pass 2: header-based matching.
    -- Claude Code overrides the tab title but shows the working directory
    -- in its header (first few lines). Read only the first 1000 chars of
    -- each pane to match against the worktree path without false positives
    -- from scrollback content.
    tell process "Ghostty"
        set wCount to count of windows
    end tell
    repeat with wIdx from 1 to wCount
        tell process "Ghostty"
            set tabCount to count of radio buttons of tab group 1 of window wIdx
        end tell
        repeat with tIdx from 1 to tabCount
            tell process "Ghostty"
                click radio button tIdx of tab group 1 of window wIdx
            end tell
            delay 0.05
            tell process "Ghostty"
                try
                    set contentArea to group 1 of group 1 of front window
                    set panes to groups of contentArea
                    repeat with pane in panes
                        try
                            set ta to text area 1 of pane
                            set taVal to value of ta
                            -- Only check the first 1000 chars (Claude Code header area).
                            if (length of taVal) > 1000 then
                                set taVal to text 1 thru 1000 of taVal
                            end if
                            if taVal contains "%s" then
                                perform action "AXRaise" of front window
                                return
                            end if
                        end try
                    end repeat
                end try
            end tell
        end repeat
    end repeat
end tell
error "no matching tab found"`, escProject, escWorktree, escRelPath)
}

// escapeAppleScript escapes a string for safe embedding in AppleScript.
// Backslashes and double quotes need escaping.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
