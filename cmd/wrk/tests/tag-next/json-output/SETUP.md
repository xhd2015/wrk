# Scenario

**Feature**: wrk --tag-next --json emits machine-readable output

```
# --json on stdout -> JSON plan/result; no ANSI color codes
wrk --tag-next --dry-run --json -> JSON only
```

## Preconditions

- JSON mode composes with `--dry-run` and `--tag-next`.

## Steps

- Descendants set `req.Args` including `--json`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```