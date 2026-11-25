# Scenario

**Feature**: `wrk --bring d1 --no-new-terminal` (no project) is rejected before any work

```
src (git) -> wrk --bring <d1> --no-new-terminal
  -> non-zero; no worktree created
  -> stderr names the flag and requires a <project> argument
```

## Steps

1. Create `src` + a valid dep.
2. Run from `src` with no project positional.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)
	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.Args = []string{"--bring", dep1, "--no-new-terminal"}
	return nil
}
```
