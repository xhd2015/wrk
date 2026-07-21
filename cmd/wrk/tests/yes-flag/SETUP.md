# Scenario

**Feature**: universal `-y` / `--yes` is a no-op synonym of default auto-yes for Y/n prompts

```
# default: all [Y/n] auto-yes without -y; -y remains accepted as synonym
# --confirm opt-in restores interactive prompts
wrk --done -y -> same as wrk --done (own ahead, cascade, non-TTY)
wrk --merge-back -y -> merge without prompt, keep worktree
wrk --set-task "new" -y -> rename without stdout-TTY requirement
wrk -y (create) -> no-op, same as bare wrk
cascade ahead on non-TTY -> -y still succeeds (auto-yes)
```

## Preconditions

- Git must be available; Go required for dep/cascade scenarios.

## Steps

- Descendants set `req.Args` with `-y` or `--yes` and assert synonym success paths.
- TTY leaves set `req.UseScriptTTY = true` (runs wrk under `script`).

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
