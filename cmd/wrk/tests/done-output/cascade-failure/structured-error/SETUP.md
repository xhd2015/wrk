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
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDivergedExternalForCascadeFail(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}  // D3: diverged cascade needs -y to reach conflict Error:
	return nil
}
```
