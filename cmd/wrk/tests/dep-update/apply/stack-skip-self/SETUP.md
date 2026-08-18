# Scenario

**Feature**: dep’s own go.mod is not pinned when the dep checkout is on the stack

```
# primary requires dep; replace dep => ./external/dep (own git, tagged)
cwd=primary -> wrk --dep-update <external/dep>
  -> pin + tidy example.com/app
  -> example.com/dep go.mod unchanged (self)
```

## Steps

1. Seed primary replace to on-stack dep checkout + file:// GOPROXY.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackSkipSelf(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
