# Scenario

**Feature**: help without `--create`/`--show` prints set-config dispatcher usage

```
wrk --set-config -h|--help
  -> dispatcher usage: actions --create / --show; pointer to action-level help
  -> exit 0, empty stderr, trailing newline
```

## Steps

- Leaves set help form only (`--help` or `-h`); no action flag.

```go
func Setup(t *testing.T, req *Request) error {
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	// Level: set-config dispatcher help (no --create / --show).
	return nil
}
```
