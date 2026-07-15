# Scenario

**Feature**: illegal flag combinations with primary (or bare invalid modifiers) are rejected

```
wrk --done|--merge-back --json -> non-zero; --json not valid with primary
wrk --push (alone)             -> non-zero; still invalid
```

## Preconditions

- `--json` is not a valid post-modifier of `--done` / `--merge-back`.
- Bare `--push` remains invalid (push requires `--tag-next` **or** a primary).

## Steps

- Leaves set the illegal combo and assert non-zero + stderr policy.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: illegal combos still run from a git main repo when useful.
	skipIfNoGit(t)
	return nil
}
```
