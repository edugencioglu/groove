package diskusage

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CmdRunner executes a command and returns its stdout.
type CmdRunner func(ctx context.Context, name string, args ...string) (string, error)

// DefaultCmdRunner runs commands via os/exec.
func DefaultCmdRunner(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w", name, args, err)
	}
	return string(out), nil
}

// ResultMsg carries the disk usage result for a single worktree back to the
// Bubbletea model.
type ResultMsg struct {
	Path  string // worktree path (key for matching)
	Usage string // human-readable size, e.g. "153M"
	Err   error
}

// ParseDuOutput extracts the size string from `du -sh <path>` output.
// The output format is: "153M\t/path/to/dir\n"
func ParseDuOutput(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}
	fields := strings.Fields(output)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// FetchOne returns a Bubbletea Cmd that runs `du -sh` for the given path
// and sends the result back as a ResultMsg.
func FetchOne(ctx context.Context, runner CmdRunner, path string) tea.Cmd {
	return func() tea.Msg {
		out, err := runner(ctx, "du", "-sh", path)
		if err != nil {
			return ResultMsg{Path: path, Err: err}
		}
		return ResultMsg{Path: path, Usage: ParseDuOutput(out)}
	}
}

// FetchAll returns a slice of Bubbletea Cmds that fetch disk usage for all
// worktree paths. Each runs independently and sends its own ResultMsg.
func FetchAll(ctx context.Context, runner CmdRunner, paths []string) []tea.Cmd {
	cmds := make([]tea.Cmd, len(paths))
	for i, p := range paths {
		cmds[i] = FetchOne(ctx, runner, p)
	}
	return cmds
}
