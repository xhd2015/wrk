# Scenario

**Feature**: --propagate-tags is mutually exclusive with other mode flags

```
# exclusive mode family
wrk --propagate-tags + peer mode flag -> non-zero, mutually exclusive
```

## Preconditions

- Peer under test for P3 is `--list` (requirement exit criterion).

## Steps

1. Leaves pair `--propagate-tags` with a conflicting mode flag.

## Context

- Same exclusive-mode family as `--projects` / `--projects-dep-graph`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves pair --propagate-tags with one peer exclusive mode flag.
	propTagsEnsureHelpersUsed()
	return nil
}
```

