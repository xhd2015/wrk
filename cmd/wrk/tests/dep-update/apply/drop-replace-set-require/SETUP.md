# Scenario

**Feature**: apply drops local replace and sets require to latest tag

```
consumer: require v0.0.1 + replace => local dep; tags v0.0.1,v0.0.2
  -> wrk --dep-update <dep>
  -> dep-update example.com/dep -> v0.0.2
  -> no replace for example.com/dep
  -> require example.com/dep v0.0.2
  -> no tidy / no go.sum
  -> exit 0
```

## Steps

1. Seed drop-replace-latest fixture.
2. Run apply.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDropReplaceLatest(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```
