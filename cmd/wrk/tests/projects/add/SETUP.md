# Scenario

**Feature**: wrk --add manually records a project

```
wrk --add <dir> -> resolve main repo -> record (source: manual) -> stdout main path
```

## Preconditions

- `wrk --add` is a standalone mode; mutually exclusive with other modes.

## Steps

- Descendants set `req.Args = []string{"--add", <path>}`.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	return nil
}
```