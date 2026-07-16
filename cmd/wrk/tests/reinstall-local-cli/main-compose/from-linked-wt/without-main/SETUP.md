# Scenario

**Feature**: without --main from linked WT plans worktree checkout modules (contrast)

```
# Exit-criteria contrast: same cwd=linked-wt, no --main
linked-wt -> wrk --reinstall-local --dry-run
  -> useMain=false → scan ShowToplevel(linked-wt)
  -> would: go install ./cmd/wtbin only (not mainbin/toolbin)
```

## Steps

1. Parent built diverged main + linked-wt; cwd = linked-wt.
2. Args = `--reinstall-local --dry-run` only (drop `--main`).
3. Expect single-module dry-run for worktree module.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
