# Scenario

**Feature**: `--pr` is an exclusive primary mode (refuse compose with main/merge-back/done/list)

```
# exclusive primary
wrk --pr … + --main | --merge-back | --done | --list
  -> non-zero
  -> stderr indicates mutual exclusion / mode conflict
```

## Steps

- Leaves seed a minimal repo and set conflicting argv (flag-layer).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	ensurePrExclusiveHelpers()
	return nil
}

// setupPrExclusiveMinimal: main repo only; enough for flag-layer exclusive checks.
func setupPrExclusiveMinimal(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
}

func ensurePrExclusiveHelpers() {
	_ = setupPrExclusiveMinimal
}
```
