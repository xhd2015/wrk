# Scenario

**Feature**: dirty **own** worktree → preflight hard error **before** cascade mutates (D2)

```
# clean external, dirty own
consumer dirty + clean external
  -> wrk --done
  -> non-zero before cascade remove
  -> external still present (must not cascade-then-fail on own)
  -> consumer still present
```

## Steps

1. Build clean contained consumer + external (replace dropped, clean).
2. Write uncommitted file on **consumer** worktree only.
3. Run bare `wrk --done`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	setupCascadePreflightCleanContained(t, req)
	writeFile(t, filepath.Join(req.WtDir, "dirty-own"), "uncommitted on consumer")
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
