# Scenario

**Feature**: create with existing-directory `<target-dir>` never writes follow-up even from home

```
# shell cwd = FakeHome (home gate would open for default create)
# second positional <target-dir> is an existing directory
FakeHome (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk <mainRepo> <targetDir> -> worktree under target with default naming; follow-up empty
```

## Steps

1. Init main repo `myrepo`.
2. Pre-create empty directory `{WorkRoot}/target` as the existing `<target-dir>`.
3. Set process cwd to FakeHome so home gate would open for default-location create.
4. Run `wrk <mainRepo> <targetDir>` with follow-up env set.
5. Expect create success under the parent with default naming and empty follow-up file.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	target := filepath.Join(req.WorkRoot, "target")
	mkdirAll(t, target)
	// Shell cwd = FakeHome: without target-dir policy, home gate would write follow-up.
	req.RepoDir = req.FakeHome
	req.UseFollowupEnv = true
	req.CLIArgs = []string{mainRepo, target}
	return nil
}
```
