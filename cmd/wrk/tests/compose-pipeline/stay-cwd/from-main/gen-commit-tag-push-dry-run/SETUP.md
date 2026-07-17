# Scenario

**Feature**: on main without done — gen-commit + tag-next + push compose (activeRoot stays main)

```
# Main with staged change + origin + root-bump lineage
myrepo (main)
  -> wrk --gen-commit-msg --commit --model=m --tag-next --push --dry-run
  -> NOT mutually exclusive
  -> order: gen-commit → tag-next → push
  -> exit 0; no real commit/tag/push
```

## Steps

1. Main+origin owned change; stage another file; baseline.
2. Run gen-commit + tag-next + push dry-run.

```go
import (
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	setupAPMainOnOrigin(t, req)

	staged := filepath.Join(req.MainRepo, "staged-for-commit.go")
	writeFile(t, staged, "package staged\n")
	runGitIsolated(t, req.MainRepo, "add", "staged-for-commit.go")

	recordAPDryRunBaseline(t, req)
	subject := strings.TrimSpace(gitOutputIsolated(t, req.MainRepo, "log", "-1", "--format=%s"))
	writeFile(t, filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", "main.head-subject"), subject+"\n")

	req.Args = []string{
		"--gen-commit-msg", "--commit", "--model=m",
		"--tag-next", "--push", "--dry-run",
	}
	return nil
}
```
