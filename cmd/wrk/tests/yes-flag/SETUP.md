# Scenario

**Feature**: universal `-y` / `--yes` auto-confirms Y/n prompts (compat with default auto-yes)

```
# -y parsed at top level; still valid for done/merge-back/set-task (default is already auto-yes)
wrk --done -y -> skip Proceed? on own ahead/diverged worktree (non-TTY ok)
wrk --merge-back -y -> merge without prompt, keep worktree
wrk --set-task "new" -y -> rename without stdout-TTY requirement
wrk -y (create) -> no-op, same as bare wrk
cascade ahead on non-TTY -> -y (and bare --done) auto-yes cascade merge
```

## Preconditions

- Git must be available; Go required for dep/cascade scenarios.

## Steps

- Descendants set `req.Args` with `-y` or `--yes` and assert prompt-skipping or guard behavior.
- TTY leaves set `req.UseScriptTTY = true` (runs wrk under `script`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
