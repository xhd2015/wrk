# Scenario

**Feature**: <target-dir> is create-only — combining it with non-create modes errors

```
# target-dir only applies to the create path; --list / --done reject it
# --bring is now allowed (relocated to create-bring/success/target-dir/)
wrk <dir> <target-dir> --list        -> wrk: unexpected arguments
```

## Preconditions

- Git must be available.

## Steps

- Leaves set `req.SpawnDir` and put the other mode in `req.Args`.
- `with-list/` uses `req.Args = ["--list"]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
