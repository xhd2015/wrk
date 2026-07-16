# Scenario

**Feature**: composed `--dir` with gen-commit + primary is rejected (wrk workDir wins)

```
# library --dir must not override primary workDir when --done is set
myrepo -> wrk --gen-commit-msg --commit --dir /tmp/other --done
  -> non-zero
  -> stderr names --dir and primary / not valid with --done
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run gen-commit + `--dir` + `--done` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	// --dir points at a dummy path; product must reject composed --dir, not use it.
	other := filepath.Join(req.WorkRoot, "other-dir")
	req.Args = []string{
		"--gen-commit-msg", "--commit", "--dir", other, "--done",
	}
	return nil
}
```
