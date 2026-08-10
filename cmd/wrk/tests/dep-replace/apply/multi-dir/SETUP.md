# Scenario

**Feature**: multi-dir apply writes absolute replaces for each dep

```
consumer + dep + dep2
  -> wrk --dep-replace <dep> <dep2>
  -> two dep-replace lines
  -> both absolute replaces in go.mod
  -> exit 0
```

## Steps

1. Seed consumer with two dep modules.
2. Run with two directory args (StringSlice multi-arg).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerTwoDeps(t, req)
	req.Args = []string{"--dep-replace", req.DepDir, req.Dep2Dir}
	return nil
}
```
