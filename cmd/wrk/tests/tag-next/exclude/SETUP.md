# Scenario

**Feature**: child-scope changes do not bump parent root scope

```
# change only under sub/ -> sub/ plans next tag; root scope skips
wrk --tag-next -> independent per-scope decisions
```

## Preconditions

- Repo has root and `sub/` scoped tags; only child-owned paths differ at HEAD.

## Steps

- Descendants use `setupSubScopeOnlyRepo` and run `--tag-next` or `--dry-run`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```