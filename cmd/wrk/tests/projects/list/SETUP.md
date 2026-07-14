# Scenario

**Feature**: wrk --projects lists detailed status for recorded projects

```
wrk --projects -> sorted detailed status blocks (one per recorded main repo)
```

## Preconditions

- `wrk --projects` is a standalone mode.

## Steps

- Descendants pre-populate `projects.json` or leave it empty.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return nil
}
```