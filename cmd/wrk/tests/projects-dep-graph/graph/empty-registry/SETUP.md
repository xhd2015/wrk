# Scenario

**Feature**: empty registry prints zero footer only

```
(no projects.json) -> wrk --projects-dep-graph
  -> exit 0
  -> stdout: "0 projects  ·  0 modules  ·  0 cross-edges\n"
  -> stderr empty
```

## Steps

1. Leave `projects.json` absent under isolated WRK_HOME.
2. Run exclusive mode (args set by graph grouping).

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	// WRK_HOME exists but projects.json is intentionally absent.
	if _, err := os.Stat(filepath.Join(req.WrkHome, "projects.json")); err == nil {
		t.Fatalf("projects.json should be absent for empty-registry leaf")
	}
	return nil
}
```
