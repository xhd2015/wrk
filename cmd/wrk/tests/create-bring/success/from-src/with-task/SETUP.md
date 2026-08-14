# Scenario

**Feature**: `wrk src -t 'fix login' --bring d1` slugs the new WT and brings into it

```
src requires dep1
  -> wrk src -t 'fix login' --no-config --bring <d1>
  -> path/branch include fix-login slug
  -> bring under that WT
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Set `req.TaskDesc = "fix login"`, `req.TaskFlag = "-t"`.
3. Run from WorkRoot (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	req.TargetDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.TaskDesc = "fix login"
	req.TaskFlag = "-t"
	req.ConsumerTop = createBringDefaultWTWithTask(req, "fix login")
	req.Args = []string{"--no-config", "--bring", dep1}
	return nil
}
```
