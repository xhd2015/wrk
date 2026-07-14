# Scenario

**Feature**: `--exec` is rejected with non-allowed modes

```
wrk --list --exec true    -> non-zero; mutual exclusion / not valid with this mode
wrk --status --exec true  -> non-zero; same class of error
# no child command run; no mode side effects required beyond error path
```

## Preconditions

- Git available for modes that require a git cwd (`--list`, `--status`).

## Steps

- Leaves set disallowed mode + `--exec` and assert non-zero + stderr.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

