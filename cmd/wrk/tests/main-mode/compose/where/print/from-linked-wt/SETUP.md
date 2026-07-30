# Scenario

**Feature**: wrk --main --where from linked worktree root prints main path

```
linked-wt (linked of myrepo) -> wrk --main --where
  -> stdout myrepo abs path\n
  -> empty stderr; no shell
```

## Steps

1. Create main repo + linked worktree.
2. Install fake bash (detect accidental launch).
3. Run `wrk --main --where` with cwd = linked worktree root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	installFakeBash(t, req, 0)
	setMainWhereArgs(req, "--main", "--where")
	return nil
}
```
