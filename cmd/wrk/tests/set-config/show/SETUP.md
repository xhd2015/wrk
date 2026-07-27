# Scenario

**Feature**: `--set-config --show` prints effective config JSON

```
wrk --set-config --show -> stdout pretty config.json; exit 0
```

## Steps

- Seed config; run show.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
