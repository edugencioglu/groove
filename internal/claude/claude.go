package claude

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/edugencioglu/groove/internal/discovery"
	"github.com/edugencioglu/groove/internal/tabs"
)

// FindClaudePIDs returns PIDs of processes whose command contains "claude".
// It filters out non-relevant matches (e.g., "claude-helper" subprocesses are
// fine — we just need any claude process whose cwd matches a worktree).
func FindClaudePIDs(procs []tabs.ProcessInfo) []int {
	var pids []int
	for _, p := range procs {
		base := filepath.Base(p.Comm)
		if base == "claude" || base == "claude-code" {
			pids = append(pids, p.PID)
		}
	}
	return pids
}

// ParsePgrepOutput parses the output of `pgrep -f claude` which returns
// one PID per line.
func ParsePgrepOutput(output string) []int {
	var pids []int
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}

// DetectClaude finds Claude Code processes and matches their cwds against
// worktree paths, setting HasClaude = true on matches. Gracefully no-ops
// if no Claude processes are found or commands fail.
func DetectClaude(ctx context.Context, runner tabs.CmdRunner, projects []discovery.Project) {
	// Try ps-based detection first (reuses the same ps output pattern as tabs).
	psOut, err := runner(ctx, "ps", "-eo", "pid,ppid,comm")
	if err != nil {
		return
	}

	procs := tabs.ParsePSOutput(psOut)
	claudePIDs := FindClaudePIDs(procs)

	// Fallback: if ps didn't find claude by comm, try pgrep which searches
	// the full command line (claude may be invoked via node/npx).
	if len(claudePIDs) == 0 {
		pgrepOut, err := runner(ctx, "pgrep", "-f", "claude")
		if err != nil {
			return // no claude processes at all
		}
		claudePIDs = ParsePgrepOutput(pgrepOut)
	}

	if len(claudePIDs) == 0 {
		return
	}

	// Build worktree path → pointer map.
	wtByPath := make(map[string]*discovery.Worktree)
	for i := range projects {
		for j := range projects[i].Worktrees {
			wtByPath[projects[i].Worktrees[j].Path] = &projects[i].Worktrees[j]
		}
	}

	// Resolve each claude process's cwd and match.
	for _, pid := range claudePIDs {
		cwd, err := tabs.ResolvePIDCwd(ctx, runner, pid)
		if err != nil {
			continue
		}
		if wt, ok := wtByPath[cwd]; ok {
			wt.HasClaude = true
		}
	}
}
