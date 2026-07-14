# Scenario

**Feature**: wrk --rm rejects invalid flag combinations

```
wrk --rm mutually exclusive with other modes
```

## Steps

- Descendants combine `--rm` with other standalone modes.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	return nil
}
```
