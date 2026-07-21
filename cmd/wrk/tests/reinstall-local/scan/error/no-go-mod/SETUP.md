# Scenario

**Feature**: no go.mod anywhere under non-git workDir → error (R4)

```
# R4: plain empty directory; no git; no go.mod on walk-up
empty/  <- workDir
  -> ResolveReinstallScanRoot / PlanLocalReinstallsFromWorkDir
  -> non-nil error (mentions go.mod)
```

## Steps

1. Create empty directory `{WorkRoot}/empty` (no go.mod, no git).
2. Set WorkDir to that path.
3. Expect non-nil error; substring includes `go.mod`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	empty := filepath.Join(req.WorkRoot, "empty")
	mkdirAll(t, empty)
	req.WorkDir = resolvePath(t, empty)
	req.WantErrSubstrs = []string{"go.mod"}
	req.WantScanRoot = ""
	req.WantModules = []WantModulePlan{}
	return nil
}
```
