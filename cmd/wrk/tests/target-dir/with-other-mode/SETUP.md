# Scenario

**Feature**: <target-dir> is create-only — combining it with other modes errors

```
# target-dir only applies to the create path; --list / --done / --dep reject it
wrk <dir> <target-dir> --list        -> wrk: unexpected arguments
wrk <dir> <target-dir> --dep <dep>   -> wrk: unexpected arguments
```

## Preconditions

- Git must be available. `with-dep` also requires Go on PATH.

## Steps

- Leaves set `req.SpawnDir` and put the other mode in `req.Args`.
- `with-list/` uses `req.Args = ["--list"]`.
- `with-dep/` builds a consumer repo requiring a dep, then `req.Args = ["--dep", depPath]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
