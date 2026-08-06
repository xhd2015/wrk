# Scenario

**Feature**: option R — `wrk --push -f` from a linked worktree force-pushes the worktree branch

```
# main has origin/main; linked wt on feature-push with unique tip
linked worktree (feature-push) + bare origin
  -> wrk --push -f   # cwd = linked wt
  -> pushed feature-push → origin/feature-push
  -> origin has feature-push tip == linked HEAD
```

## Steps

1. Seed main + bare origin; create linked worktree on `feature-push` with a commit.
2. Run `wrk --push -f` with cwd = linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushFromLinkedWorktree(t, req)
	req.Args = []string{"--push", "-f"}
	return nil
}
```
