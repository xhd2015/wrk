# Scenario

**Feature**: `wrk --new -y` create mode ( -y is no-op on create)

```
main repo -> wrk --new -y -> worktree path on stdout; same side effects as wrk --new
```

## Steps

1. Create main repo.
2. Run `wrk --new -y` from main repo.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	req.RepoDir = mainRepo
	req.Args = []string{"--new", "-y"}
	return nil
}
```
