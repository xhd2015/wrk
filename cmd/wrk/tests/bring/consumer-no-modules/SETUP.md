# Scenario

**Feature**: wrk --bring soft-skips replace when consumer has zero Go modules

```
# consumer git with zero go.mod files -> wrk --bring
#   -> exit 0; external worktree + /external gitignore
#   -> no replace; SKIP consumer has no Go modules (distinct wording)
consumer (git, no go.mod) + mydep (module example.com/dep)
  -> wrk --bring <dep>
  -> stdout external path; stderr SKIP … consumer has no Go modules
```

## Preconditions

- Git and Go available.
- Consumer repo root has NO go.mod and no subdirectory go.mod.
- Consumer cwd is the repo root.

## Steps

1. Create consumer git repo with no go.mod anywhere.
2. Create dep git repo with valid go.mod.
3. Run `wrk --bring <dep>` from the consumer.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	// NO go.mod at repo root. No go.mod in any subdirectory.
	writeFile(t, filepath.Join(consumer, "README.md"), "# consumer\n")
	runGitIsolated(t, consumer, "add", "README.md")
	runGitIsolated(t, consumer, "commit", "-m", "init consumer")

	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--bring", dep}
	return nil
}
```
