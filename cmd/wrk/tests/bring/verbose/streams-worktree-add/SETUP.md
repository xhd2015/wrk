# Scenario

**Feature**: wrk --bring -v streams git worktree add output (Preparing worktree)

```
# new external create with -v -> stderr has worktree add pre-line + git progress
# mirrors fetch-and-verbose/verbose/create/streams-output
consumer (require dep) + mydep -> wrk --bring <dep> -v
  -> stderr: timestamp git worktree add + Preparing worktree / HEAD is now at
```

## Steps

1. Matching fixtures (first bring → always creates new external).
2. Run with `-v`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep, "-v"}
	return nil
}
```
