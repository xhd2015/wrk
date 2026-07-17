# Scenario

**Feature**: hard rejects for illegal compose combinations (independent of activeRoot)

```
# --done + --merge-back still exclusive
# multi-stage + --json rejected (--json only with bare --tag-next)
wrk --done --merge-back / wrk --sync --tag-next --json
  -> non-zero clear stderr
```

## Steps

- Grouping only.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
