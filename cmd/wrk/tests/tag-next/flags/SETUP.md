# Scenario

**Feature**: wrk --tag-next flag validation and mutual exclusion

```
# invalid flag combos -> non-zero exit before tagscope runs
wrk --dry-run (alone) / --tag-next --done -> errors
```

## Preconditions

- `--dry-run` requires `--all-deps` or `--tag-next`.
- `--tag-next` is mutually exclusive with `--done` and other modes.

## Steps

- Descendants set `req.Args` for the invalid combination under test.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```