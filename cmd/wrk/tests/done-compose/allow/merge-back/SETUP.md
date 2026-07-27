# Scenario

**Feature**: `--merge-back` primary accepts composition modifiers at flag layer

```
wrk --merge-back [--gen-commit-msg --commit …] [--tag-next|…] from main
  -> not mutually exclusive with allowed modifiers
```

## Steps

- Leaves under this node always include `--merge-back` plus allowed modifiers
  (including P2 `--gen-commit-msg --commit` pre-stage).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Narrow primary to --merge-back; leaves add modifiers on main-repo cwd.
	skipIfNoGit(t)
	return nil
}
```
