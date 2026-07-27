# Scenario

**Feature**: P2 non-TTY bare `wrk` prints a static fine-grained dashboard snapshot

```
# linked or main checkout, no args, non-TTY
cwd (git) -> wrk
  -> exit 0
  -> stdout: dashboard View (Pre / Main / After / Batch)
  -> glyphs [•] / [ ] / [-] only (no [x])
  -> stages: add changes, gen-commit-msg + nested agent-runner (commandcode),
             commit, MERGE BACK then DONE, sync, tag-next, push, reinstall-local
  -> MERGE BACK default [•] when linked; no create-hint
  -> Batch would-run line present
  -> no create under WRK_HOME/worktrees
  -> event command "dashboard" (see events-command-dashboard leaf)

# not required in P2
# interactive Bubble Tea keys / RUN compose / agent-runner cycling
```

## Steps

- Leaves set cwd (linked dirty/clean or main) and empty Args.
- Assertions use shared `assertDashboardSnapshotCore` / `assertAddChangesGlyph` / `assertMergeBackDefaultSelected`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	// Snapshot leaves are bare no-args only.
	req.Args = nil
	req.TargetDir = ""
	req.TaskDesc = ""
	return nil
}
```
