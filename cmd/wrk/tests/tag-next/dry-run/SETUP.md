# Scenario

**Feature**: wrk --tag-next --dry-run plans tags without creating git refs

```
# tagscope Plan only -> stdout per-scope lines + N tag planned; no git tag
wrk --tag-next --dry-run -> plan summary, no side effects
```

## Preconditions

- Git available; repo has recognizable version tags and owned-file changes (or none).

## Steps

- Descendants seed a tagged repo and set `req.Args = []string{"--tag-next", "--dry-run"}`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```