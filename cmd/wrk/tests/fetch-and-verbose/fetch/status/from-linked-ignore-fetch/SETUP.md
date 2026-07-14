# Scenario

**Feature**: --fetch ignored on linked worktree cwd (no fetch, no Remote:)

```
linked wt cwd + --status --fetch -> same stdout as without --fetch; stderr empty
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "linked-main")
	initFetchVerboseRepo(t, mainRepo, "linked main base")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-branch")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--status", "--fetch"}
	return nil
}
```