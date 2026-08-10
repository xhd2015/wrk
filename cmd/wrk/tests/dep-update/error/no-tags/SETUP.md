# Scenario

**Feature**: --dep-update fails when dep has no version tags

```
consumer has replace; dep is go module git repo without tags
  -> wrk --dep-update <dep>
  -> non-zero; go.mod unchanged
```

## Steps

1. Seed no-tags fixture.
2. Run update.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupNoTagsDep(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
