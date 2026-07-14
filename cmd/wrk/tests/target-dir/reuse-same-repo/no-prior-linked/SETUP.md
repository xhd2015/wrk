# Scenario

**Feature**: named bring with no prior linked worktree of source creates as today (no skip prompt)

```
# no live linked WT of myrepo -> wrk myrepo <target> creates new WT under target
# no "already exists" / skip prompt on stderr
myrepo (main only) -> wrk myrepo {WorkRoot}/target -> {WorkRoot}/target/myrepo-main-{date}
```

## Steps

1. Ensure source `myrepo` has no linked worktrees (parent init only).
2. Pre-create `{WorkRoot}/target` as an existing directory.
3. Set `req.SpawnDir = {WorkRoot}/target`.
4. Run `wrk myrepo <target>` via `Run`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureNamedBringReuseHelpersUsed()
	target := filepath.Join(req.WorkRoot, "target")
	mkdirAll(t, target)
	req.SpawnDir = target
	return nil
}
```
