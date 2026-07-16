# Scenario

**Feature**: single registered project with one module and no cross edges

```
projects.json = [solo]
solo: example.com/solo (dir=.)
wrk --projects-dep-graph
  -> project block + module line
  -> no "→" cross-edge lines
  -> footer 1 project · 1 module · 0 cross-edges
```

## Steps

1. Create `repos/solo` with root module `example.com/solo` (no requires).
2. Register its absolute path in projects.json.
3. Run exclusive mode.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	solo := filepath.Join(req.WorkRoot, "repos", "solo")
	initSingleModuleRepo(t, solo, "example.com/solo")
	solo = resolvePath(t, solo)
	req.SoloPath = solo
	writeProjectsJSON(t, req.WrkHome, solo)
	return nil
}
```
