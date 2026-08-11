# Scenario

**Feature**: wrk --dep-update with neither dirs nor --all errors

```
wrk --dep-update
  -> non-zero
  -> requires directory or --all
```

## Steps

1. Run `wrk --dep-update` with no path args and no `--all`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--dep-update"}
	return nil
}
```
