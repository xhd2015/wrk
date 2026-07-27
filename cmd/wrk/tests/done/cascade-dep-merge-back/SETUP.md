# Scenario

**Feature**: bare non-TTY `--done` auto-yes merges cascaded ahead external dep (then replace guard)

```
# external dep wt ahead of dep main; default auto-yes runs cascade before replace guard
consumer wt + ahead external/dep wt -> wrk --done
  -> cascade merges dep + removes external wt
  -> parent errors on local replace (go.mod still points at external)
```

## Steps

1. Build consumer wt with ahead external dep via `setupConsumerWithAheadExternalDep`.
2. Run bare `wrk --done` from the consumer worktree (non-TTY). Leave replace in place.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}  // D3: cascade not-included needs -y
	return nil
}
```
