# Scenario

**Feature**: `--exec` cut-flag parse errors (no trailing args; equals form)

```
wrk --exec              -> non-zero; requires a command / requires arguments
wrk --exec=pwd          -> non-zero; cut is not a value flag (equals form rejected)
# Prefer fail at parse before create / mode work
```

## Preconditions

- No special git fixture required when parse fails first; leaves may use WorkRoot as cwd.

## Steps

- Leaves set bare or equals-form `--exec` and assert non-zero + stderr.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
