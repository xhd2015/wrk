# Scenario

**Feature**: wrk --rm rejects invalid flag combinations

```
wrk --rm mutually exclusive with other modes
```

## Steps

- Descendants combine `--rm` with other standalone modes.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```
