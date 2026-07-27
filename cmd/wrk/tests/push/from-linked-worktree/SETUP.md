# Scenario

**Feature**: option R — bare `wrk --push` from a linked worktree pushes the worktree branch

```
# main has origin/main; linked wt on feature-push with unique tip
linked worktree (feature-push) + bare origin
  -> wrk --push   # cwd = linked wt
  -> pushed feature-push → origin/feature-push
  -> origin has feature-push tip == linked HEAD
  -> does NOT require pushing main (option R: current checkout branch)
```

## Steps

1. Seed main + bare origin; create linked worktree on `feature-push` with a commit.
2. Run `wrk --push` with cwd = linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushFromLinkedWorktree(t, req)
	req.Args = []string{"--push"}
	return nil
}
```
