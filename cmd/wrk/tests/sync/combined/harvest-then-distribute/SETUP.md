# Scenario

**Feature**: One run harvests an ahead worktree into main then distributes to a behind worktree

```
# base C0 on main
# feature-login: +2 clean commits (ahead)
# feature-api: still at C0 (behind after harvest)
myrepo + login + api -> wrk --sync
  -> main ← feature-login  (+2 commits)
  -> feature-api ← main  (+2 commits)
  -> synced: 1 into main, 1 into worktrees, 0 skipped
```

## Steps

1. Init main with init commit.
2. Add linked worktree `feature-login` first, then `feature-api` (list order).
3. On `feature-login` only: two clean commits (`feat: one`, `feat: two`).
4. Leave `feature-api` at base (behind).
5. Run `wrk --sync` from main.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	commitFile(t, mainRepo, "README.md", "# sync fixture\n", "init")
	mainRepo = resolveRepoPath(t, mainRepo)

	// Create harvest candidate first so ListLinked order is stable: login then api.
	wtLogin := addLinkedWorktree(t, mainRepo, "feature-login", filepath.Join(req.WorkRoot, "wt-feature-login"))
	wtAPI := addLinkedWorktree(t, mainRepo, "feature-api", filepath.Join(req.WorkRoot, "wt-feature-api"))

	commitFile(t, wtLogin, "a.txt", "one\n", "feat: one")
	loginTip := commitFile(t, wtLogin, "b.txt", "two\n", "feat: two")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtLogin
	req.WtBranch = "feature-login"
	req.WtSHA = loginTip
	req.Wt2Path = wtAPI
	req.Wt2Branch = "feature-api"
	req.Wt2SHA = revParseHEAD(t, wtAPI)
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.Args = []string{"--sync"}
	return nil
}
```
