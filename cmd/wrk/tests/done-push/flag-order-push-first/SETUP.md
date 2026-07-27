# Scenario

**Feature**: flag order free — `--push --done` same as `--done --push`

```
# --push before --done still composes; origin receives main tip after merge
myrepo (origin) + wt (feature-work)
  -> wrk --push --done -y
  -> same stdout/side effects as --done -y --push
```

## Steps

1. Same origin + ahead worktree fixture as `pushes-main`.
2. Run `wrk --push --done -y` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePushWithOrigin(t, req)
	req.Args = []string{"--push", "--done", "-y"}
	return nil
}
```
