# Scenario

**Feature**: wrk --sync rejects positional arguments

```
wrk --sync some-path
  -> non-zero exit
  -> stderr: unexpected arguments
  -> stdout empty
```

## Steps

1. `initMainOnlyRepo` (valid git cwd so the error is about args, not git).
2. Run `wrk --sync some-path` from the main repo.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync", "some-path"}
	return nil
}
```
