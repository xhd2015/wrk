# Scenario

**Feature**: hard rejects for illegal compose combinations (independent of successful activeRoot switch)

```
# --done + --merge-back still exclusive
# multi-stage + --json rejected (--json only with bare --tag-next)
# --main exclusive with --done / --merge-back / --gen-commit-msg
wrk --done --merge-back / wrk --sync --tag-next --json
wrk --main --done | --main --merge-back | --main --gen-commit-msg …
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
