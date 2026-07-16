# Scenario

**Feature**: wrk --gen-commit-msg --no-verify requires --commit

```
# flag validation before agent / dry-run success
workspace/ -> wrk --gen-commit-msg --no-verify
  -> non-zero; message about requires --commit
```

## Preconditions

- No git repository required; validation fails at flag parse / library entry.

## Steps

1. Run `wrk --gen-commit-msg --no-verify` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--gen-commit-msg", "--no-verify"}
	return nil
}
```
