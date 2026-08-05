# Scenario

**Feature**: --done follow-up after successful worktree remove

```
linked wt + WRK_FOLLOWUP_FILE
wrk --done -> follow-up: cd <main-repo-abs>

# foreign parent (different git main within 3 levels) -> no follow-up
# --force-cd bypasses cwd-missing and foreign-repo gates
sibling A (cwd); wrk --done B --force-cd + env -> cd <main>
sibling A; wrk --done B --force-cd (no channel) -> shell @ main
```

## Steps

1. Descendants create linked worktree then run --done.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "binary")
	return nil
}
```
