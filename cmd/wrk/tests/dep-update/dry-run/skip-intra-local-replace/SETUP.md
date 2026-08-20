# Scenario

**Feature**: dir-mode dry-run reflects intra-module replace skip

```
# primary requires dep + replace dep => ./external/dep (own git, on the stack)
# dep has nested cmd sub-module: require dep + replace dep => ../ (intra-module)
cwd=primary -> wrk --dep-update <dep> --dry-run
  -> would: pin + would: go mod tidy on primary (checkout .)
  -> would: skip on cmd sub-module (checkout external/dep)
  -> both go.mods unchanged; no go.sum
```

## Steps

1. Seed the same stack topology as the apply variant.
2. Run dry-run from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDirModeSkipIntraReplace(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```
