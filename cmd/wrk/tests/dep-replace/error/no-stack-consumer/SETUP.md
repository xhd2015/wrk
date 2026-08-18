# Scenario

**Feature**: git stack with zero gated consumers errors and prints no banner

```
# primary + external/kool neither require nor replace dep
cwd=primary -> wrk --dep-replace <dep>
  -> wrk: error containing replace or consumer
  -> no ==== banner
  -> both go.mods unchanged
```

## Steps

1. Seed git stack with no require/replace of dep.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackZeroGated(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
