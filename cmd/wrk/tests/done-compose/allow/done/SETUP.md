# Scenario

**Feature**: `--done` primary accepts composition modifiers at flag layer

```
wrk --done [--gen-commit-msg --commit …] [--tag-next|--push|--sync|--reinstall-local|--dry-run...] from main
  -> not mutually exclusive / not only-valid-with-tag-next
```

## Steps

- Leaves under this node always include `--done` plus one or more allowed modifiers
  (post stages and/or P2 `--gen-commit-msg --commit` pre-stage).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Narrow primary to --done; leaves add modifiers on main-repo cwd.
	skipIfNoGit(t)
	return nil
}
```
