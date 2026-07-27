# Scenario

**Feature**: Edge cases — main detached fatal; detached linked worktree skip

```
# main detached HEAD -> wrk --sync -> fatal Error (non-zero)
# linked wt detached HEAD -> warning skip; exit 0
```

## Steps

- Descendants build detached fixtures.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
