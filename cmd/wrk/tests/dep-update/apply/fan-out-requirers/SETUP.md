# Scenario

**Feature**: dir-mode pins every existing requirer under git toplevel

```
# git workspace: example.com/app and pkg/ example.com/app/pkg both require dep
cwd=workspace -> wrk --dep-update <dep>
  -> checkout . ; pin + tidy root app then pkg
  -> both requires @v0.0.2; both go.sum exist
  -> updated 2 modules in 1 checkouts
```

## Steps

1. Seed git workspace with two requirers + file:// GOPROXY.
2. Run apply from the git toplevel.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFanOutRequirers(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
