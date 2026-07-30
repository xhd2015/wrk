# Scenario

**Feature**: wrk --where --main (partner first) same success as --main --where

```
linked-wt -> wrk --where --main
  -> stdout main abs\n; empty stderr; no shell
```

## Steps

1. Create main + linked worktree; cwd = linked root.
2. Args = `--where`, `--main` (order under test).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	installFakeBash(t, req, 0)
	setMainWhereArgs(req, "--where", "--main")
	return nil
}
```
