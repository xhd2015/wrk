# Scenario

**Feature**: two positionals `wrk <dir> <arg2>` without `-t` — arg2 may be task-like or a real target-dir

```
wrk <dir> <arg2> -> classify arg2 (task-like vs path-like vs short token)
  -> task-like: confirm / -y / non-TTY error
  -> not task-like: unchanged target-dir create
```

## Steps

- Leaves set `TargetDir` + `SpawnDir` from WorkRoot shell cwd.
- Do not set `TaskDesc` unless testing explicit `-t` coexistence.

```go
func Setup(t *testing.T, req *Request) error {
	// Repo assembled per-leaf via setupTwoArg / initMyrepoForForgotTask.
	skipIfNoGit(t)
	return nil
}
```
