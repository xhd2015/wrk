# Scenario

**Feature**: two dep args; one stack consumer gets both pins + one tidy; the other requires only the first

```
# primary requires dep + dep2 + kool; kool requires only dep
cwd=primary -> wrk --dep-update <dep> <dep2>
  -> two dep headers (argv order)
  -> app: both pins + one tidy
  -> kool: one pin + one tidy
```

## Steps

1. Seed two tagged deps + stack other-checkout + file:// GOPROXY.
2. Run multi-arg apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMultiDirStack(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir, req.Dep2Dir}
	return nil
}
```
