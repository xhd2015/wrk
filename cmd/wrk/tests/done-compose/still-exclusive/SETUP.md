# Scenario

**Feature**: non-composed mode pairs remain mutually exclusive after composition lands

```
# composition must not accidentally open other exclusives
wrk --tag-next --list -> still mutually exclusive
```

## Preconditions

- Standalone exclusives such as `--tag-next` + `--list` stay rejected.

## Steps

- Leaves set still-invalid mode pairs.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: regression exclusives need a valid git cwd for mode flags.
	skipIfNoGit(t)
	return nil
}
```
