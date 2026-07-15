# Scenario

**Feature**: `--done` primary accepts composition modifiers at flag layer

```
wrk --done [--tag-next|--push|--sync|--dry-run...] from main
  -> not mutually exclusive / not only-valid-with-tag-next
```

## Steps

- Leaves under this node always include `--done` plus one or more allowed modifiers.

```go
func Setup(t *testing.T, req *Request) error {
	// Narrow primary to --done; leaves add modifiers on main-repo cwd.
	skipIfNoGit(t)
	return nil
}
```
