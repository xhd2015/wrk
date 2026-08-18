# Scenario

**Feature**: multi-dir dry-run prints two dep headers and writes nothing

```
# primary requires dep+dep2; kool requires only dep
cwd=primary -> wrk --dep-replace <dep> <dep2> --dry-run
  -> two dep headers
  -> would: replace; no write
```

## Steps

1. Seed multi-dir stack fixture.
2. Run dry-run with both dep dirs.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMultiDirStack(t, req)
	req.Args = []string{"--dep-replace", req.DepDir, req.Dep2Dir, "--dry-run"}
	return nil
}
```
