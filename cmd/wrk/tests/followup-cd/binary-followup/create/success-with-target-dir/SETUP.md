# Scenario

**Feature**: create with explicit `<target-dir>` (missing, parent exists) never writes follow-up even from home

```
# shell cwd = FakeHome (home gate would open for default create)
# second positional <target-dir> missing; parent exists
FakeHome (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk <mainRepo> <target> -> worktree exactly at <target>; follow-up empty
```

## Steps

1. Init main repo `myrepo`.
2. Choose absolute `<target>` under WorkRoot that does not exist (parent WorkRoot exists).
3. Set process cwd to FakeHome so home gate would open for default-location create.
4. Run `wrk <mainRepo> <target>` with follow-up env set.
5. Expect create success at exact target path and empty follow-up file.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	// Absolute target: missing path whose parent (WorkRoot) exists.
	target := filepath.Join(req.WorkRoot, "wt")
	// Shell cwd must be exact user home so a home-gated write would fire without target-dir policy.
	req.RepoDir = req.FakeHome
	req.UseFollowupEnv = true
	req.CLIArgs = []string{mainRepo, target}
	return nil
}
```
