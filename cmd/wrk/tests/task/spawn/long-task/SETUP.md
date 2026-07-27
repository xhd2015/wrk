# Scenario

**Feature**: --task with >64 runes is truncated to 64

```
wrk --task "a very long task description..." -> slug truncated at 64 runes
```

## Steps

1. Create repo.
2. Set req.TaskDesc with 80 runes.
3. Verify slug in dir and branch names is at most 64 runes.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	req.TaskDesc = "explore the integration of distributed tracing with opentelemetry across all microservices"
	req.RepoDir = mainRepo
	expectedSlug := slugify(req.TaskDesc)
	req.WtBranch = branchNameWithTask("main", wrkDate, expectedSlug, 0)
	return nil
}
```