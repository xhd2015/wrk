# Scenario

**Feature**: `wrk . --bring d1 --no-new-terminal` is create+bring (project positional `.`)

```
# the fix for the silent-create bug: a project positional keeps create+bring valid,
# so "." is the documented way to spell "create from here".
src (git) -> wrk . --bring <d1> --no-new-terminal
  -> new default WT; bring into that WT (not into src)
  -> event command=create; args include --bring and d1
```

## Steps

1. Create `src` requiring dep1 + a valid `mydep1`.
2. Run from inside `src` with `.` as the project positional (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{".", "--bring", dep1, "--no-new-terminal"}
	return nil
}
```
