# Scenario

**Feature**: `--here` treats `-v` / `--verbose` as unset (no git verbose chatter)

```
src requires dep1
  -> wrk src --no-config --here -v --bring <d1>
  -> create path only on stdout
  -> stderr has no timestamped git worktree add / Preparing worktree stream
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Run with `--here -v --bring` (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	req.TargetDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--no-config", "--here", "-v", "--bring", dep1}
	return nil
}
```
