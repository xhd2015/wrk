# Scenario

**Feature**: wrk --all-deps silently skips projects.json entries whose path does not exist

```
# projects.json lists nonexistent path -> skip silently -> wrked 0 deps
consumer (requires dep1) + projects.json (missing path) -> wrked 0 deps
```

## Steps

1. Create a consumer requiring `example.com/dep1`.
2. Seed `projects.json` with a path that does not exist on disk.
3. Run `wrk --all-deps` from the consumer.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")
	missing := filepath.Join(req.WorkRoot, "does-not-exist", "mydep1")
	writeAllDepsProjectsJSON(t, req.WrkHome, missing)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```