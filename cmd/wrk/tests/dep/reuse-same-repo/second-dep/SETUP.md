# Scenario

**Feature**: second `wrk --dep` of the same required dep reuses the first external worktree and keeps replace

```
# first --dep creates external + replace
# second --dep same dep -> same path; reuse warning; replace still correct; no -1
consumer (require dep) + mydep
  -> wrk --dep mydep (1st) -> external/mydep-main-{date}
  -> wrk --dep mydep (2nd) -> reuse same path + replace
```

## Steps

1. Create consumer requiring `example.com/dep` and dep repo `mydep`.
2. Run `wrk --dep <dep>` once; record external path.
3. Run `wrk --dep <dep>` again via doctest `Run`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepReuseHelpersUsed()

	consumer := initConsumerRepo(t, req.WorkRoot, true)
	depPath := initDepRepo(t, req.WorkRoot, "mydep", true)
	// Canonicalize dep like bring fixtures (path comparisons / messages).
	if resolved, err := filepath.EvalSymlinks(depPath); err == nil {
		depPath = resolved
	}

	first := runWrkWithArgs(t, req, consumer, "--dep", depPath)
	wantFirst := externalWorktreePath(consumer, "mydep", "main", 0)
	if first != wantFirst {
		t.Fatalf("first --dep: expected %q, got %q", wantFirst, first)
	}

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePath
	req.ExternalWtDir = wantFirst
	req.Args = []string{"--dep", depPath}
	return nil
}
```
