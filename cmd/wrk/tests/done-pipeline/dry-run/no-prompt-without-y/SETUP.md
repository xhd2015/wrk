# Scenario

**Feature**: dry-run does not hang or require confirm without `-y` (non-TTY)

```
# ahead worktree + non-TTY stdin (default doctest pipe)
myrepo + wt (ahead)
  -> wrk --done --dry-run
  -> exit 0 promptly with plan; no Proceed?; no "stdin is not a terminal"
  -> no mutations
```

## Steps

1. Root-bump seed + linked worktree ahead (NeedsConfirm relation).
2. Snapshot baseline.
3. Run `wrk --done --dry-run` with empty stdin (no `-y`, no `--confirm-from-stdin`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineLocal(t, req)
	recordComposeDryRunBaseline(t, req)
	// Explicit empty stdin; non-TTY. Must not require -y.
	req.Args = []string{"--done", "--dry-run"}
	req.StdinInput = ""
	return nil
}
```
