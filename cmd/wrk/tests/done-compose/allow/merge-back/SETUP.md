# Scenario

**Feature**: `--merge-back` primary accepts composition modifiers at flag layer

```
wrk --merge-back [--tag-next|…] from main
  -> not mutually exclusive with allowed modifiers
```

## Steps

- Leaves under this node always include `--merge-back` plus allowed modifiers.

```go
func Setup(t *testing.T, req *Request) error {
	// Narrow primary to --merge-back; leaves add modifiers on main-repo cwd.
	skipIfNoGit(t)
	return nil
}
```
