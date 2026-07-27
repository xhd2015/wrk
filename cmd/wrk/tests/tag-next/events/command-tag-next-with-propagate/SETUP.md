# Scenario

**Feature**: --tag-next --propagate-tags records events.jsonl command "tag-next" (primary)

```
# compose flags still log primary tag-next (not propagate-tags)
# go.mod required so propagate-tags stage finds source modules
git module repo -> wrk --tag-next --propagate-tags --dry-run -> event command=tag-next
```

## Steps

1. Init a single-module git repo with lightweight release tag and post-tag owned change.
2. Run `wrk --tag-next --propagate-tags --dry-run` (plan only; event command still tag-next).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// setupRootBumpRepo has no go.mod; propagate-tags needs scanned modules.
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/lib\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repo, "lib.go"), "package lib\n")
	runGitIsolated(t, repo, "add", ".")
	runGitIsolated(t, repo, "commit", "-m", "init module")
	createLightweightTag(t, repo, "v0.0.1", "")
	commitFile(t, repo, "lib.go", "package lib // bump\n", "post-tag change")
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--tag-next", "--propagate-tags", "--dry-run"}
	return nil
}
```
