# Scenario

**Feature**: wrk --dry-run flag validation (without --all-deps)

```
# bare --dry-run without a host mode -> error before any planning
wrk --dry-run -> error (host list includes --all-deps, --done, --merge-back, --tag-next, --sync)
```

## Preconditions

- Bare `--dry-run` (no host) is rejected; stderr still contains `--dry-run is only valid with --all-deps` as a substring of the full host list.
- Other hosts (`--done` / `--merge-back` / `--tag-next` / `--sync`) are covered outside this all-deps subtree.
- Registered-project dry-run planning leaves live under `registered/dry-run/`.

## Steps

- Descendants invoke `wrk --dry-run` without `--all-deps` and assert stderr.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()
	return nil
}
```