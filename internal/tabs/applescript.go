package tabs

import (
	"fmt"
	"strings"
)

// NewTabScript generates an AppleScript that opens a new Ghostty tab in the
// given directory using Ghostty's native scripting API (1.3+).
func NewTabScript(path string) string {
	escaped := escapeAppleScript(path)
	return fmt.Sprintf(`tell application "Ghostty"
    activate
    set cfg to new surface configuration
    set initial working directory of cfg to "%s"
    new tab in front window with configuration cfg
end tell`, escaped)
}

// JumpTabScript generates an AppleScript that switches to an existing Ghostty
// tab whose terminal's working directory matches the given absolute path.
// Uses Ghostty's native scripting API (1.3+) to enumerate terminals and match
// by working directory.
//
// Returns an AppleScript error if no matching tab is found.
func JumpTabScript(path string) string {
	escaped := escapeAppleScript(path)
	return fmt.Sprintf(`tell application "Ghostty"
    activate
    set winCount to count of windows
    repeat with wIdx from 1 to winCount
        set w to window wIdx
        set allTabs to every tab of w
        repeat with t in allTabs
            set termList to every terminal of t
            repeat with term in termList
                if working directory of term is "%s" then
                    focus term
                    return
                end if
            end repeat
        end repeat
    end repeat
end tell
error "no matching tab found"`, escaped)
}

// escapeAppleScript escapes a string for safe embedding in AppleScript.
// Backslashes and double quotes need escaping.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
