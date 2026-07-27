# Scenario

**Feature**: --status from main subdirectory (≤2 ups) uses Rel Dir, not forced `.`

```
# nested cwd under main: two levels (pkg/api)
cwd = main/pkg/api
wrk --status -> main Dir: ../..  (+ Remote); external Dir via statusDirLine
```

## Steps

1. Create main + one external wrk worktree.
2. Create nested dirs `pkg/api` under main (no need to commit).
3. Run `wrk --status` with cwd = `main/pkg/api`.

## Context

- Pure Rel: do **not** force `Dir: .` merely because cwd is inside the checkout.
- Leading `..` count for main is 2 → relative `../..` OK.
- External under `{WorkRoot}/.wrk/worktrees/…` from `main/pkg/api` typically needs
  three `..` → **absolute** under the rule.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, _, _ := createExternalWrkWorktree(t, req)
	subdir := filepath.Join(mainRepo, "pkg", "api")
	mkdirAll(t, subdir)
	req.RepoDir = subdir
	return nil
}
```
