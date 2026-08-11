# Scenario

**Feature**: --all soft warnings (exit 0) for no-tag / missing registry soft-skips

```
inventory owner without numeric tags
  -> wrk --dep-update --all
  -> warning: on stderr; skip; exit 0
```

## Steps

1. Leaves seed soft-fail fixtures and set Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	if len(req.Args) == 0 {
		req.Args = []string{"--dep-update", "--all"}
	}
	ensureDepUpdateHelpersUsed()
	return nil
}
```
