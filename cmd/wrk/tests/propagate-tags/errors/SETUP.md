# Scenario

**Feature**: hard errors for --propagate-tags (invalid cwd, no tags, mutual exclusion)

```
# non-git | no source numeric tags | exclusive peer flags
wrk --propagate-tags [...] -> exit ≠ 0, clear stderr
```

## Preconditions

- These leaves expect non-zero exit.
- Success dry-run plan leaves live under sibling `dry-run/`.

## Steps

1. Grouping marks hard-error scenarios.
2. Leaves set failing preconditions and args.

## Context

- Stdout empty for mutual exclusion; other errors may leave stdout empty too.

```go
func Setup(t *testing.T, req *Request) error {
	// Hard-error subtree; leaves set failing preconditions and args.
	propTagsEnsureHelpersUsed()
	return nil
}
```

