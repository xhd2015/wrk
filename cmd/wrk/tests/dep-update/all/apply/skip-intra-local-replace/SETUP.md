# Scenario

**Feature**: --all skips same-toplevel filesystem replace deps; still bumps inventory

```
# mono requires mono/lib with replace => ./lib (skip) + lib@v1.0.0 inventory (bump)
cwd=mono -> wrk --dep-update --all
  -> dep-update example.com/lib -> v1.2.3
  -> skipped count includes local replace
  -> replace for mono/lib still present; require mono/lib unchanged
```

## Steps

1. Seed monorepo with local replace + inventory owner require; modproxy.
2. Run apply from mono root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllSkipIntraLocalReplace(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```
