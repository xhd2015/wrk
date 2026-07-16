# Scenario

**Feature**: bare wrk --dry-run rejected; host list includes --propagate-tags

```
# wrk --dry-run alone -> validation error listing hosts
# hosts must include --propagate-tags after P3 lands
git cwd -> wrk --dry-run -> non-zero stderr with --propagate-tags
```

## Steps

1. Create a minimal git repo as cwd (flag validation may run after git checks
   or before — either way bare dry-run must fail).
2. Run `wrk --dry-run` without any host mode.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initSingleModuleRepo(t, repo, "example.com/tmp", nil)
	repo = resolvePath(t, repo)
	req.RepoDir = repo
	req.SourcePath = repo
	req.Args = []string{"--dry-run"}
	return nil
}
```
