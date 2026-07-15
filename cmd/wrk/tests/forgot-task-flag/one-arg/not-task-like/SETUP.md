# Scenario

**Feature**: one-arg that resolves as source — no treat-as-task prompt

```
wrk <existing-repo-path|basename>
  -> normal create; never promote-as-task
```

## Steps

- Leaves provide a resolvable source positional.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: leaves configure positionals and interactive mode.
	skipIfNoGit(t)
	return nil
}
```
