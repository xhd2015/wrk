# Scenario

**Feature**: `--set-config --create` writes create UX keys

```
wrk --set-config --create <flags> -> config.json create.* merged; exit 0
```

## Steps

- Leaves choose positive/negative flags under `--create`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
