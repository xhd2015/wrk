# Scenario

**Feature**: full pre + primary + posts + reinstall tail with `--done` passes flag validation (P3 integration)

```
# P3 full ship at flag layer: gen-commit pre + done + sync/tag-next/push + reinstall tail
myrepo -> wrk --gen-commit-msg --commit --model=m --done --sync --tag-next --push --reinstall-local -y
  -> not mutually exclusive
  -> not "--push is only valid with --tag-next" false reject
  -> may later fail: not a linked worktree / no staged changes (flag-layer only)
```

## Preconditions

- P1 unlocked reinstall tail with posts; P2 unlocked gen-commit pre with posts.
- This leaf proves **all bands together** at the flag matrix (pre + primary + posts + tail).
- Full ordered plan lives under `done-pipeline/dry-run/full-combo-gen-commit-reinstall/`.

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run full pre+post+tail combo (with `-y`) from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{
		"--gen-commit-msg", "--commit", "--model=m",
		"--done", "--sync", "--tag-next", "--push",
		"--reinstall-local", "-y",
	}
	return nil
}
```
