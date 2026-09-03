# Scenario

**Feature**: compose soft-skip for empty-index `--gen-commit-msg --commit`

```
# with a later pipeline stage: soft-skip clean tree
wrk --add-all --gen-commit-msg --commit --exec true
  -> notice: worktree clean, skip commit; exit 0

# bare (no partner): still hard-fail
wrk --add-all --gen-commit-msg --commit
  -> non-zero; no staged
```

## Preconditions

- Leaves use hooks-disabled isolated git under WorkRoot.
- Prefer L2 `InProcess` Capture (no agent needed for empty-index paths).

## Steps

1. Parent Setup is a no-op beyond git availability.
2. Leaves init a clean repo and set Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
