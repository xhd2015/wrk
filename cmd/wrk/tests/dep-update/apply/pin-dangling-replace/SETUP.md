# Scenario

**Feature**: dir-mode pins a dep whose filesystem replace target is dangling (absent), instead of skipping it as intra-module

```
# primary (git) requires dep + replace dep => ./external/dep (target absent)
cwd=primary -> wrk --dep-update <dep>
  -> pin + tidy example.com/app (checkout .)
  -> no "skip ... (intra-module replace)"
  -> go.mod: replace dropped, require bumped to v0.0.2; go.sum exists
```

## Steps

1. Seed the same topology as the dry-run variant.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDirModePinDanglingReplace(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
