# Scenario

**Feature**: preferred external path occupied (not a same-repo worktree) → path `-1`; branch stays preferred when free

```
# external/mydep exists as a plain dir; preferred branch main-{date} free
consumer --bring mydep
  -> path external/mydep-1
  -> branch main-{date} (path and branch suffixes independent; no same-repo Policy A reuse)
```

## Steps

1. Create consumer requiring `example.com/dep`.
2. Create dep repo `mydep` on `main`.
3. Pre-create plain directory `{consumer}/external/mydep` (not a git worktree).
4. Run `wrk --bring <dep>` from consumer.

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	depPath := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	blocked := bringExternalWorktreePath(consumer, "mydep", "main", 0)
	if err := os.MkdirAll(blocked, 0o755); err != nil {
		t.Fatalf("mkdir blocked path: %v", err)
	}
	// Marker so the dir is clearly not a worktree.
	if err := os.WriteFile(filepath.Join(blocked, "not-a-worktree"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```
