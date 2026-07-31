# Scenario

**Feature**: `--push --pr` always full-pushes the worktree branch, then runs the PR path

```
# remote head present or missing; local tip may be ahead
linked wt + github origin + fake gh
  -> wrk --push --pr --title T --comment C
  -> full branch push (updates tip even when remote head already exists)
  -> then PR create (or attach + comment)
  -> stdout: pushed … then PR tokens
```

## Steps

- Leaves choose remote-present (local ahead) vs remote-missing fixtures.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
