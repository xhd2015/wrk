# Scenario

**Feature**: other-checkout module without require of xxx is left unchanged (default quiet)

```
# primary requires dep + kool; replace kool => ./external/kool
# kool does not require dep
cwd=primary -> wrk --dep-update <dep>
  -> pin + tidy example.com/app only
  -> kool go.mod unchanged; no skip line
```

## Steps

1. Seed stack with non-requirer other checkout + file:// GOPROXY.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackSkipNonRequirerOther(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
