# Scenario

**Feature**: wrk --cd CLI forms `wrk --cd PATH` and `wrk PATH --cd` are equivalent

```
# Bool("--cd") + exactly one positional path
wrk --cd /abs  <->  wrk /abs --cd
# same resolve + in-place mode (channel open; no hang)
```

## Preconditions

- Uses in-place channel so both forms assert follow-up without needing a shell.
- Same absolute target directory for both siblings.

## Steps

1. Open in-place channel; create abs target.
2. Leaf chooses flag-then-path vs path-then-flag.

## Context

- Both leaves expect exit 0, empty stdout, follow-up `cd /abs\n`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	enableInPlaceChannel(t, req)
	target := cdAbsTarget(t, req, "jumpto")
	req.MainRepo = target // reuse field as resolved abs target for asserts
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	return nil
}
```
