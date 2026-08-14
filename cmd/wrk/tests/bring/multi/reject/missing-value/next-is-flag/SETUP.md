# Scenario

**Feature**: `wrk --bring --no-dep` treats the flag as not a value

```
# next token starts with - → Varargs does not consume it
wrk --bring --no-dep -> non-zero; requires a value
```

## Steps

1. Run `wrk --bring --no-dep` from isolated WorkRoot (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--bring", "--no-dep"}
	return nil
}
```
