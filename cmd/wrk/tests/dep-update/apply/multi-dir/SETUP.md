# Scenario

**Feature**: multi-dir apply updates each dep (drop replace + require@latest)

```
consumer replace+require for dep and dep2; both tagged
  -> wrk --dep-update <dep> <dep2>
  -> two dep-update lines with latest versions
  -> go mod tidy ok once for example.com/consumer
  -> both replaces dropped
  -> exit 0
```

## Steps

1. Seed two tagged deps + file:// GOPROXY.
2. Run multi-arg update.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoTaggedDeps(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir, req.Dep2Dir}
	return nil
}
```
