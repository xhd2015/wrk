# Scenario

**Feature**: `--gen-commit-msg --commit --dry-run --done` plans would-commit then would-merge; zero mutations (P2 pre-stage dry-run)

```
# staged change on linked wt + done dry-run composition
myrepo (v0.0.1) + wt (feature-work ahead) + staged uncommitted file
  -> wrk --gen-commit-msg --commit --dry-run --done
  -> gen-commit dry plan: mock B / would: git commit (no real commit)
  -> primary dry plan: MergeBack DryRun (ff-merge + remove + branch -D)
  -> HEAD subject + main tip + wt link unchanged
  -> no mutually exclusive at flag layer
```

## Preconditions

- Parent fixtures: `setupDonePipelineLocal` + dry-run baseline helpers.
- Classic RED until gen-commit peels as pre-stage before primary (today early path mutexes `--done`).

## Steps

1. Root-bump seed + wrk-managed linked worktree ahead (`feature-work`).
2. Stage one additional text file on the worktree (uncommitted index change).
3. Snapshot dry-run baseline SHAs + pre-run HEAD subject on the worktree.
4. Run `wrk --gen-commit-msg --commit --dry-run --done` from the worktree (no `-y`).

```go
import (
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	setupDonePipelineLocal(t, req)

	// Stage an uncommitted change on the source worktree for gen-commit dry plan.
	staged := filepath.Join(req.WtDir, "staged-for-commit.go")
	writeFile(t, staged, "package staged\n")
	runGitIsolated(t, req.WtDir, "add", "staged-for-commit.go")

	recordComposeDryRunBaseline(t, req)

	// Record subject before dry-run so Assert can prove no commit.
	subject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	writeFile(t, filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", "wt.head-subject"), subject+"\n")

	req.Args = []string{"--gen-commit-msg", "--commit", "--dry-run", "--done"}
	return nil
}
```
