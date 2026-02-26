package shell

import (
	"fmt"
	"os"
)

// ZshHook returns the zsh shell integration snippet. The snippet installs a
// precmd hook that sets the Ghostty tab title to the output of `groove title`.
func ZshHook() string {
	grooveBin, err := os.Executable()
	if err != nil {
		grooveBin = "groove"
	}

	return fmt.Sprintf(`# groove shell integration — auto-sets Ghostty tab title to project/worktree
_groove_precmd() {
  local title
  title=$(%s title 2>/dev/null)
  if [[ -n "$title" ]]; then
    printf '\e]2;%%s\a' "$title"
  fi
}
precmd_functions+=(_groove_precmd)
`, grooveBin)
}
