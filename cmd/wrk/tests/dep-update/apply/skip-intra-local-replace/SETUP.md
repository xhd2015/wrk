# Scenario

**Feature**: dir-mode skips intra-module replace deps on a stack checkout

```
# primary requires dep + replace dep => ./external/dep (own git, on the stack)
# dep has nested cmd sub-module: require dep + replace dep => ../ (intra-module)
cwd=primary -> wrk --dep-update <dep>
  -> pin + tidy example.com/app (checkout .)
  -> skip example.com/dep/cmd (checkout external/dep) — intra-module replace
  -> cmd go.mod unchanged; primary replace dropped, require bumped
```

## Steps

1. Seed stack with dep's own git repo containing a cmd sub-module with
   intra-module replace `dep => ../`.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDirModeSkipIntraReplace(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
