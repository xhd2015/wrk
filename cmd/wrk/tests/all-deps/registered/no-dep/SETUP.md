# Scenario

**Feature**: wrk --all-deps --no-dep links registered deps as worktrees without replace/tidy

```
# consumer requires dep1+dep2; both registered -> wrk --all-deps --no-dep
#   -> external wts + wrked N
#   -> consumer go.mod has NO new replaces
consumer (requires dep1, dep2) + projects.json
  -> wrk --all-deps --no-dep
  -> wrk example.com/depN at ./external/... ; wrked 2 deps; no replaces
```

## Steps

1. Create and register `mydep1` / `mydep2`.
2. Create consumer requiring both modules; snapshot go.mod.
3. Run `wrk --all-deps --no-dep`.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	registeredEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	initAllDepsRepo(t, dep2, "example.com/dep2", "dep2")
	registerAllDepsProjects(t, req, dep1, dep2)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	// Snapshot go.mod for byte-identity after --no-dep.
	data, err := os.ReadFile(filepath.Join(consumer, "go.mod"))
	if err != nil {
		t.Fatalf("snapshot go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(req.WorkRoot, "go.mod.before"), data, 0o644); err != nil {
		t.Fatalf("write go.mod.before: %v", err)
	}

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--no-dep"}
	return nil
}
```
