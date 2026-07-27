# Scenario

**Feature**: `wrk --new --no-config` still performs native create

```
# --no-config is a create modifier; with --new remains create
myrepo -> wrk --new --no-config
  -> exit 0; stdout worktree path\n
  -> worktree under WRK_HOME
```

## Steps

1. Init main repo (parent).
2. Optionally leave config absent/empty (no create.* needed for this leaf).
3. Run `wrk --new --no-config`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--new", "--no-config"}
	return nil
}
```
