# Scenario

**Feature**: `wrk src dst --bring d1` spawns at missing `dst` and brings into that path

```
# parent exists; dst missing → exact spawn path (no naming suffix)
src requires dep1
  -> wrk src {WorkRoot}/dst --no-config --bring <d1>
  -> worktree at dst; external/ under dst
  -> src/go.mod unchanged; src/external missing
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Snapshot `src/go.mod`.
3. Set `SpawnDir = {WorkRoot}/dst` (must not exist).
4. Run from WorkRoot (L2).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	createBringSnapshotGoMod(t, req, src)
	dst := filepath.Join(req.WorkRoot, "dst")
	req.TargetDir = src
	req.SpawnDir = dst
	req.MainRepo = src
	req.DepPath = dep1
	req.ConsumerTop = dst
	req.Args = []string{"--no-config", "--bring", dep1}
	return nil
}
```
