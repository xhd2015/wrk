# Scenario

**Feature**: diverged external cascade merge-back fails with structured `Error:` including path

```
# main-side vs external-side conflict on dep.go
consumer + diverged external
  -> wrk --done
  -> ==> cascade  (targets ≥ 1)
  -> cascade MergeBack fails (rebase conflict under the hood)
  -> stderr: Error: … + path context for external worktree
  -> ==> own may be absent
  -> consumer own MergeBack must not complete (wt still present)
```

## Steps

1. Build consumer + external dep with diverged `dep.go` via `setupDivergedExternalForCascadeFail`.
2. Run bare `wrk --done` from consumer worktree (default auto-yes; non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupDivergedExternalForCascadeFail(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
