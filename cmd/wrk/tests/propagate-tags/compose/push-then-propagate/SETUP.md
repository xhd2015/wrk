# Scenario

**Feature**: compose with --push publishes branch+tags then propagates consumers

```
# root-bump lib + bare origin; app requires older lib
cwd=lib -> wrk --tag-next --propagate-tags --push
  -> (1) tag-next apply: create v1.0.1 locally
  -> (2) push branch tip + v1.0.1 to origin + confirm line
  -> blank line
  -> (3) propagate apply: bump app to v1.0.1 + build + commit
  -> origin has refs/heads/main + refs/tags/v1.0.1; app require at v1.0.1
```

## Steps

1. `setupComposeRootBump` with origin (bare remote + push main).
2. Args: `--tag-next --propagate-tags --push` (order free; push is mid-stage after tag-next).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	setupComposeRootBump(t, req, true)
	// Advance source after origin was seeded so tags-only push cannot satisfy
	// "branch tip on origin" (origin lags source HEAD). Re-capture snapshots.
	writeFile(t, filepath.Join(req.SourcePath, "lib.go"),
		"package lib\n\nfunc Version() string { return \""+req.NextTag+"-tip\" }\n")
	runGitIsolated(t, req.SourcePath, "add", "lib.go")
	runGitIsolated(t, req.SourcePath, "commit", "-m", "local tip after origin seed")
	// Proxy was seeded from previous HEAD; re-seed next version from current tree.
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, req.ModulePath, req.NextTag, req.SourcePath)
	captureRepoSnapshots(t, req)
	req.Args = []string{"--tag-next", "--propagate-tags", "--push"}
	return nil
}
```
