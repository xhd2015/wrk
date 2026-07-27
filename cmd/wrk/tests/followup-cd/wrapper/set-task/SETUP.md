# Scenario

**Feature**: wrapper --set-task auto-cds to new path after move

```
source bash.sh from old wt; wrk --set-task "new"
  -> stderr cd <newPath>; FinalPWD = newPath
```

## Steps

1. Descendants create task worktree and set --set-task args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "wrapper")
	return nil
}
```
