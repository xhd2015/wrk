# Scenario

**Feature**: wrk --web rejects positional arguments

```
wrk --web some-dir
  -> non-zero exit
  -> stderr: unexpected arguments
  -> stdout empty
```

## Steps

1. Run `wrk --web some-dir` from isolated WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--web", "some-dir"}
	return nil
}
```
