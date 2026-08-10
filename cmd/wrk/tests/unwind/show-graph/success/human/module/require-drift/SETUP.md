# Scenario

**Feature**: require version ≠ dep latest tag → human `(latest …)` drift

```
# sibling follow: require v0.0.0; dep tagged v0.0.1
wrk --unwind --show-graph
  -> edge shows require v0.0.0 and (latest v0.0.1)
  -> may also show replaced (local filesystem replace)
```

## Steps

1. Seed sibling local-replace stack (root + dot-pkgs, both dirty).
2. Tag dep main at `v0.0.1` so tagscope latest differs from require `v0.0.0`.
3. Run show-graph.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowSiblingBothDirty(t, req)
	// Dep latest tag v0.0.1; consumer still requires v0.0.0 → drift.
	if req.DepsLinkedWtDir == "" {
		t.Fatal("require-drift: missing dep checkout")
	}
	createLightweightTag(t, req.DepsLinkedWtDir, unwindApplyOldTag, "HEAD")
	req.OldRequireVersion = "v0.0.0"
	req.ExpectedPinVersion = unwindApplyOldTag // latest on dep
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
