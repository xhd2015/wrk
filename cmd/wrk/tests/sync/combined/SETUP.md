# Scenario

**Feature**: Combined pass1 harvest then pass2 distribute in one `wrk --sync`

```
# wt1 ahead of main; wt2 behind main
# pass1 harvests wt1 into main; pass2 distributes new main into wt2
main + feature-login + feature-api -> wrk --sync -> both directions
```

## Steps

- Descendants build the two-worktree fixture.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
