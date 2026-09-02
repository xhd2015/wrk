# Scenario

**Feature**: `--here` + multi `--bring` suppresses bring chatter

```
src requires dep1+dep2
  -> wrk src --no-config --here --bring <d1> <d2>
  -> stdout is create path only
  -> stderr has no will bring: / SKIP
  -> both externals exist under the new WT; replaces applied
```

## Steps

1. Create `src` requiring both modules + `mydep1` / `mydep2`.
2. Run with `--here --bring` from WorkRoot (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module, createBringDep2Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)
	dep2 := initCreateBringDep(t, req.WorkRoot, "mydep2", createBringDep2Module)

	createBringSnapshotGoMod(t, req, src)
	req.TargetDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.SecondRepo = dep2
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--no-config", "--here", "--bring", dep1, dep2}
	return nil
}
```
