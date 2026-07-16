# Scenario

**Feature**: CLI dry-run without go.mod yields a clear non-zero error

```
# C6: ModuleRoot exists, no go.mod; ancestors under WorkRoot also have none
mod/ -> wrk --reinstall-local --dry-run -> non-zero, error mentions go.mod
```

## Steps

1. Leave ModuleRoot empty (root Setup mkdir'd it; no go.mod written).
2. Run `wrk --reinstall-local --dry-run` from ModuleRoot.

```go
func Setup(t *testing.T, req *Request) error {
	// ModuleRoot exists but has no go.mod — C6.
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
