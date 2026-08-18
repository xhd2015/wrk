# Scenario

**Feature**: --all dry-run prints the stack tree for an other-checkout inventory require

```
# primary + external/kool both require example.com/lib@v1.0.0
cwd=primary -> wrk --dep-update --all --dry-run
  -> ==== dep-update (dry-run) ====
  -> no argv dep list; would: pin on both checkouts
  -> zero writes
```

## Steps

1. Seed stack + registered owner.
2. Run `--all --dry-run` from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllStackOutdated(t, req)
	req.Args = []string{"--dep-update", "--all", "--dry-run"}
	return nil
}
```
