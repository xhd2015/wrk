# Scenario

**Feature**: universal `-y` / `--yes` auto-confirms Y/n prompts on wrk CLI

```
# -y parsed at top level; auto-yes on own-worktree merge-back and --set-task rename
wrk --done -y -> skip Proceed? on own ahead/diverged worktree (non-TTY ok)
wrk --merge-back -y -> merge without prompt, keep worktree
wrk --set-task "new" -y -> rename without stdout-TTY requirement
wrk -y (create) -> no-op, same as bare wrk
cascade ahead on non-TTY -> -y ineffective; TTY + -y auto-confirms cascade
```

## Preconditions

- Git must be available; Go required for dep/cascade scenarios.

## Steps

- Descendants set `req.Args` with `-y` or `--yes` and assert prompt-skipping or guard behavior.
- TTY leaves set `req.UseScriptTTY = true` (runs wrk under `script`).

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
