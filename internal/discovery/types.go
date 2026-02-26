package discovery

// Project represents a git repository that may contain multiple worktrees.
type Project struct {
	Name      string
	Path      string // absolute path to the main repo
	Worktrees []Worktree
}

// Worktree represents a single git worktree within a project.
type Worktree struct {
	Name   string // derived from directory name
	Path   string // absolute path
	Branch string // current branch or "(detached)" or "(bare)"

	// Populated by gitstatus
	IsClean   bool
	Modified  int
	Untracked int
	Ahead     int
	Behind    int
	TabCount  int // number of Ghostty tabs open for this worktree
	HasClaude bool
	DiskUsage string
}
