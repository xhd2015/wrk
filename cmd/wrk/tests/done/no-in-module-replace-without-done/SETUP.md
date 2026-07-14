# Scenario

**Feature**: --no-in-module-replace is rejected unless combined with --done

```
# --no-in-module-replace only meaningful for --done's guard; with --list it must error
wrk --list --no-in-module-replace -> non-zero, "only valid with --done"
```

## Steps

1. Create a main repo and a linked worktree via `wrk` (so `--list` would
   otherwise be valid from the worktree).
2. Run `wrk --list --no-in-module-replace` from the worktree.

## Expected (correct) behavior

`--no-in-module-replace` is valid only with `--done`. Combined with any other
mode (`--list`, `--dep`, `--all-deps`, no-args create) it errors with
`wrk: --no-in-module-replace is only valid with --done` and a non-zero exit,
mirroring `--confirm-from-stdin`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	req.RepoDir = wtDir
	req.Args = []string{"--list", "--no-in-module-replace"}
	return nil
}
```
