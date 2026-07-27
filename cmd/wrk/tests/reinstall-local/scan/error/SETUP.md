# Scenario

**Feature**: scan-root resolution fails when no root can be found

```
# unresolvable workDir: not in git and no go.mod on walk-up
workDir (empty / no go.mod, non-git)
  -> ResolveReinstallScanRoot / PlanLocalReinstallsFromWorkDir
  -> error
```

## Preconditions

- Leaves under this branch set `WantError=true`.
- No git init on the WorkDir tree.

## Steps

1. Group marks error expectations default.
2. Leaves create empty / go.mod-free directories and set WantErrSubstrs.

## Context

- Error text should mention `go.mod` (or the failed walk) so callers can
  surface a clear CLI message later.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseMain = false
	req.WantError = true
	return nil
}
```
