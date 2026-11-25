# Scenario

**Feature**: `wrk --new --bring d1` (no project) is rejected — use `wrk <project> --new --bring d1`

```
src (git) -> wrk --new --bring <d1>
  -> non-zero; no worktree created
  -> stderr names --new and requires a <project> argument
```

## Steps

1. Create `src` + a valid dep.
2. Run from `src` with `--new` but no project positional.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)
	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.Args = []string{"--new", "--bring", dep1}
	return nil
}
```
