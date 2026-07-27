# Scenario

**Feature**: wrk --dry-run flag validation (without --all-deps)

```
# bare --dry-run without a host mode -> error before any planning
wrk --dry-run -> error (host list: done|merge-back|all-deps|tag-next|propagate-tags|sync)
```

## Preconditions

- Bare `--dry-run` (no host) is rejected; stderr lists hosts including `--all-deps` and `--propagate-tags`.
- Other hosts (`--done` / `--merge-back` / `--tag-next` / `--propagate-tags` / `--sync`) are covered outside this all-deps subtree.
- Registered-project dry-run planning leaves live under `registered/dry-run/`.

## Steps

- Descendants invoke `wrk --dry-run` without `--all-deps` and assert stderr.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	allDepsEnsureHelpersUsed()
	return nil
}
```