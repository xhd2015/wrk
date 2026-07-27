# Scenario

**Feature**: wrapper --done auto-cds to main repo

```
source bash.sh from linked wt; wrk --done
  -> stderr cd <main>; FinalPWD = main
```

## Steps

1. Descendants prepare already-included linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "wrapper")
	return nil
}
```
