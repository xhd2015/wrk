# Scenario

**Feature**: single-dir apply writes absolute replace

```
consumer requires example.com/dep; no replace
  -> wrk --dep-replace <dep>
  -> dep-replace example.com/dep => <abs>
  -> go.mod replace absolute (not relative)
  -> no go.sum / no tidy
  -> exit 0
```

## Steps

1. Seed consumer with require + dep module.
2. Run apply for single dep dir.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
