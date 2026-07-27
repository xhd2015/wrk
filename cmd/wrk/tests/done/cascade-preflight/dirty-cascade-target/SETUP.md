# Scenario

**Feature**: dirty cascade target → preflight hard error before any mutation (D2)

```
# dirty external linked wt, clean own
consumer + dirty external
  -> wrk --done
  -> non-zero; Error: + dirty/uncommitted language + external path
  -> external still present; consumer still present
```

## Steps

1. Build clean contained consumer + external.
2. Write uncommitted file on external worktree only.
3. Run bare `wrk --done`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadePreflightCleanContained(t, req)
	writeFile(t, filepath.Join(req.ExternalWtDir, "dirty-ext"), "uncommitted on cascade target")
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
