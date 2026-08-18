# Scenario

**Feature**: dir-mode pins a requirer on another stack checkout found via replace BFS

```
# primary requires dep + kool; replace kool => ./external/kool (own git)
# kool already requires dep
cwd=primary -> wrk --dep-update <dep>
  -> pin + tidy example.com/app (checkout .)
  -> pin + tidy example.com/kool (checkout external/kool)
  -> both requires @v0.0.2
```

## Steps

1. Seed primary + independent `external/kool` git repo + filesystem replace + file:// GOPROXY.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackOtherCheckoutRequirer(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
