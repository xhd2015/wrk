# Scenario

**Feature**: project modes reject invalid flag combinations

```
wrk --projects / --add mutually exclusive with other modes
wrk --add requires a path argument
```

## Steps

- Descendants combine invalid flag sets or omit required `--add` path.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```