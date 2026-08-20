# Scenario

**Feature**: dir-mode dry-run pins a dep whose filesystem replace target is dangling (absent), instead of skipping it as intra-module

```
# primary (git) requires dep + replace dep => ./external/dep (target absent)
cwd=primary -> wrk --dep-update <dep> --dry-run
  -> would: pin + would: go mod tidy on primary (checkout .)
  -> no "would: skip ... (intra-module replace)"
  -> go.mod unchanged; no go.sum
```

## Steps

1. Seed a tagged dep repo (v0.0.1, v0.0.2).
2. Seed a git primary whose go.mod requires dep v0.0.1 and replaces dep with a
   relative `./external/dep` that does not exist on disk.
3. Run dry-run from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDirModePinDanglingReplace(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```
