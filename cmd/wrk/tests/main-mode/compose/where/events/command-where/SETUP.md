# Scenario

**Feature**: successful wrk --main --where records events.jsonl command "where"

```
linked-wt -> wrk --main --where
  -> exit 0
  -> last event: command=where, exit_code=0, args include --main and --where
```

## Steps

1. Create main + linked worktree; cwd = linked root.
2. Run `wrk --main --where`.

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
