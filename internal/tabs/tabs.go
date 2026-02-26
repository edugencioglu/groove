package tabs

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/edugencioglu/groove/internal/discovery"
)

// CmdRunner executes a command and returns its stdout. Injected for testing.
type CmdRunner func(ctx context.Context, name string, args ...string) (string, error)

// DefaultCmdRunner executes commands via os/exec.
func DefaultCmdRunner(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w", name, args, err)
	}
	return string(out), nil
}

// ProcessInfo holds a parsed process entry from ps output.
type ProcessInfo struct {
	PID  int
	PPID int
	Comm string
}

// ParsePSOutput parses the output of `ps -eo pid,ppid,comm` into ProcessInfo
// entries. The first line (header) is skipped.
func ParsePSOutput(output string) []ProcessInfo {
	var procs []ProcessInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue // skip header or malformed lines
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		comm := fields[2]
		procs = append(procs, ProcessInfo{PID: pid, PPID: ppid, Comm: comm})
	}
	return procs
}

// FindGhosttyShellPIDs returns PIDs of shell processes that are children of
// Ghostty. It walks the process tree: finds ghostty PIDs, then finds
// intermediate processes (like the login shell launcher), then finds shells
// (zsh, bash, fish) that are descendants.
func FindGhosttyShellPIDs(procs []ProcessInfo) []int {
	// Build parent→children map.
	children := make(map[int][]int)
	procByPID := make(map[int]ProcessInfo)
	for _, p := range procs {
		children[p.PPID] = append(children[p.PPID], p.PID)
		procByPID[p.PID] = p
	}

	// Find ghostty PIDs.
	var ghosttyPIDs []int
	for _, p := range procs {
		base := filepath.Base(p.Comm)
		if base == "ghostty" {
			ghosttyPIDs = append(ghosttyPIDs, p.PID)
		}
	}

	if len(ghosttyPIDs) == 0 {
		return nil
	}

	// BFS to find all descendant shell processes.
	shellNames := map[string]bool{
		"zsh": true, "bash": true, "fish": true,
		"-zsh": true, "-bash": true, "-fish": true,
	}

	var shellPIDs []int
	var queue []int
	for _, gp := range ghosttyPIDs {
		queue = append(queue, children[gp]...)
	}

	visited := make(map[int]bool)
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		if visited[pid] {
			continue
		}
		visited[pid] = true

		p := procByPID[pid]
		base := filepath.Base(p.Comm)
		if shellNames[base] {
			shellPIDs = append(shellPIDs, pid)
		}
		// Continue BFS through children in case shells are deeper.
		queue = append(queue, children[pid]...)
	}

	return shellPIDs
}

// ParseLsofCwd parses the output of `lsof -a -p <pid> -d cwd -Fn` and
// returns the current working directory. The output looks like:
//
//	p12345
//	fcwd
//	n/path/to/directory
func ParseLsofCwd(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "n/") {
			return line[1:] // strip the leading 'n'
		}
	}
	return ""
}

// ResolvePIDCwd gets the current working directory for a process via lsof.
func ResolvePIDCwd(ctx context.Context, runner CmdRunner, pid int) (string, error) {
	out, err := runner(ctx, "/usr/sbin/lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn")
	if err != nil {
		return "", err
	}
	cwd := ParseLsofCwd(out)
	if cwd == "" {
		return "", fmt.Errorf("no cwd found for pid %d", pid)
	}
	return cwd, nil
}

// DetectTabs finds Ghostty shell processes and matches their cwds against
// the given worktree paths, incrementing TabCount on matches. If Ghostty
// isn't running, this is a no-op (graceful degradation).
func DetectTabs(ctx context.Context, runner CmdRunner, projects []discovery.Project) {
	psOut, err := runner(ctx, "ps", "-eo", "pid,ppid,comm")
	if err != nil {
		return // ps failed — can't detect tabs
	}

	procs := ParsePSOutput(psOut)
	shellPIDs := FindGhosttyShellPIDs(procs)
	if len(shellPIDs) == 0 {
		return // Ghostty not running or no shells
	}

	// Build worktree path → pointer map for fast lookup.
	wtByPath := make(map[string]*discovery.Worktree)
	for i := range projects {
		for j := range projects[i].Worktrees {
			wtByPath[projects[i].Worktrees[j].Path] = &projects[i].Worktrees[j]
		}
	}

	// Resolve each shell's cwd and match.
	for _, pid := range shellPIDs {
		cwd, err := ResolvePIDCwd(ctx, runner, pid)
		if err != nil {
			continue
		}
		if wt, ok := wtByPath[cwd]; ok {
			wt.TabCount++
		}
	}
}
