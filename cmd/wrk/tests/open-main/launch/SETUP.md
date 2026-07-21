# Scenario

**Feature**: wrk --main when not already at main root launches nested LoginInteractive in main repo

```
# channel always nested for --main (ignore WRK_FOLLOWUP_FILE)
cwd != mainRepo
wrk --main
  -> shell/interactive.LoginInteractive(mainRepo, Base(mainRepo), "WRK_SHELL=1")
  -> minimal UX: no install hint; no stdout path
  -> wrk exit = shell exit
```

## Preconditions

- Every successful launch leaf **must** call `installFakeBash` so CI cannot hang.
- Follow-up env is **not** set except `followup-ignored/` (which proves ignore).
- Descendants place cwd in main subdir, linked worktree root, or linked worktree subdir.

## Steps

1. Create main repo (+ optional linked worktree / subdirs).
2. Install fake interactive shell (exit 0 unless overridden).
3. Run `wrk --main` with process cwd set by leaf.

## Context

- Shell cwd must be the **main repo root**, never the linked worktree path or subdir.
- Minimal launch UX differs from `--cd` Branch B (no stderr install hint, no stdout abs path).

```go
func Setup(t *testing.T, req *Request) error {
	// Default CLI form for all launch leaves; leaves may re-set via setMainArgs.
	setMainArgs(req)
	return nil
}
```
