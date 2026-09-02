# Scenario

**Feature**: `--here` suppresses soft SKIP when a brought dep is not required

```
src requires dep1 only
  -> wrk src --no-config --here --bring <d1> <d2>
  -> stdout create path only
  -> stderr has no SKIP local dep replacement (d2 soft-skip silenced)
  -> both externals still exist
```

## Steps

1. Create `src` requiring only dep1; create both dep repos.
2. Run with `--here --bring` both (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)
	dep2 := initCreateBringDep(t, req.WorkRoot, "mydep2", createBringDep2Module)

	req.TargetDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.SecondRepo = dep2
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--no-config", "--here", "--bring", dep1, dep2}
	return nil
}
```
