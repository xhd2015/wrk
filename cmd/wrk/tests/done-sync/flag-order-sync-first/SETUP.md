# Scenario

**Feature**: flag order free — `--sync --done` same as `--done --sync`

```
# --sync before --done still composes; wtB receives pass2 after done
myrepo + wtA (ahead) + wtB (feature-stays)
  -> wrk --sync --done -y
  -> same stdout/side effects as --done -y --sync
```

## Steps

1. Same two-worktree fixture as `basic-propagate`.
2. Run `wrk --sync --done -y` from wtA.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCompositionTwoWTs(t, req)
	req.Args = []string{"--sync", "--done", "-y"}
	return nil
}
```
