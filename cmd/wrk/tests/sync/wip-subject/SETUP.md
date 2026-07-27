# Scenario

**Feature**: `wrkcli.IsWipSubject` pure helper — WIP commit-subject prefixes

```
# after trim, case-insensitive prefixes:
#   wip:  |  wip(  |  [wip]
# empty / whitespace-only / mid-string only → false
subject string -> IsWipSubject -> bool
```

## Preconditions

- Package `github.com/xhd2015/wrk/wrkcli` is importable from this module.
- `IsWipSubject` is a pure function; no git, no CLI binary, no `WRK_HOME`.
- Root `Run` dual-mode: `req.WipProbe == true` probes the helper and sets
  `resp.IsWip` without invoking the wrk binary.

## Steps

1. Enable pure-function probe mode for all descendants.
2. Leaves set `req.Subject` to the commit subject under test.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.WipProbe = true
	return nil
}
```
