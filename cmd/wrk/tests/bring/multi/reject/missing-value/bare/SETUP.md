# Scenario

**Feature**: `wrk --bring` with no value is a parse error

```
wrk --bring -> non-zero; library wording: requires a value
```

## Steps

1. Run `wrk --bring` from isolated WorkRoot (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--bring"}
	return nil
}
```
