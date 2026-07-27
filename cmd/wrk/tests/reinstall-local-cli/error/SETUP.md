# Scenario

**Feature**: reinstall-local CLI errors when cwd is not under a Go module

```
# no go.mod in cwd or ancestors -> non-zero clear error
mod/ (empty) -> wrk --reinstall-local --dry-run -> error
```

## Preconditions

- Leaves leave ModuleRoot without a valid `go.mod` (and WorkRoot has none).
- Process cwd is ModuleRoot so walk-up cannot find a module.

## Steps

1. Leaves arrange missing go.mod.
2. Run `--reinstall-local --dry-run`.
3. Assert non-zero exit and clear error.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: module-resolution error path.
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
