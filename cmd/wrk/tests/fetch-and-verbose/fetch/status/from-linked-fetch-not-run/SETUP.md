# Scenario

**Feature**: --fetch not executed from linked worktree cwd even with -v

```
linked wt cwd + --status --fetch -v -> stderr has no fetch line
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "linked-v-main")
	initFetchVerboseRepo(t, mainRepo, "linked v main base")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-branch")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--status", "--fetch", "-v"}
	return nil
}
```