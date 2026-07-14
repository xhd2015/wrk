# Scenario

**Feature**: wrk --status from linked worktree cwd omits Remote: everywhere

```
linked wt cwd -> blocks have Master: where applicable; no Remote:
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "from-linked")
	initFetchVerboseRepo(t, mainRepo, "from linked base")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-from-linked")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--status"}
	return nil
}
```