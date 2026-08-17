# Scenario

**Feature**: sibling module under the same git toplevel that does not require xxx is left alone

```
# git workspace: app/ requires dep; other/ does not
cwd=workspace -> wrk --dep-update <dep>
  -> pin + tidy example.com/app
  -> example.com/other go.mod unchanged
```

## Steps

1. Seed git workspace with requirer + non-requirer sibling + file:// GOPROXY.
2. Run apply from the git toplevel.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSkipNonRequirer(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
