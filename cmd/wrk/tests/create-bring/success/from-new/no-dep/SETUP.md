# Scenario

**Feature**: create + `--bring --no-dep` materializes external without replace

```
src cwd -> wrk --new --no-config --bring <d1> --no-dep
  -> new WT + external; new WT go.mod unchanged (no replace)
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Snapshot `src/go.mod` (source must stay unchanged; new WT starts as a copy).
3. Run `--new --bring d1 --no-dep`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	createBringSnapshotGoMod(t, req, src)
	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--new", "--no-config", "--bring", dep1, "--no-dep"}
	return nil
}
```
