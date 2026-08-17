# Scenario

**Feature**: dir-mode errors when no scanned module already requires the dep

```
# git repo with go.mod that does not require example.com/dep
cwd=workspace -> wrk --dep-update <dep>
  -> non-zero wrk: error containing requires
  -> go.mod unchanged
```

## Steps

1. Seed tagged dep + git consumer module that does not require it.
2. Run update from the git toplevel.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupNoConsumerRequires(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
