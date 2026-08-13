## Expected

- Exit code 0.
- Combined stdout/stderr must **not** contain `does not contain package` or
  `unknown revision`.
- Combined output must **not** pin mid ← free @ **old** LatestTag
  (`pin skills <- dot-pkgs @ v0.0.1`) after the leaf has new unpublished
  source.
- **Tag-before-pin:** `tag-next example.com/dot-pkgs @ v0.0.2` **before**
  `pin skills <- dot-pkgs @ v0.0.2`.
- Free main after apply (`req.SecondRepo`):
  - `color/color.go` present (unpublished WIP landed)
  - local tag `v0.0.2` exists
- Mid after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace for that module
  - FEATURE_WIP landed when present

## Side Effects

- Today RED: pinReady on early mid peel pins free @ LatestTag after leaf
  land; tidy drops replace and dies `@latest found (v0.0.1), but does not
  contain package example.com/dot-pkgs/color`.
- Desired: land → tag-next free → pin mid @ next → then mid feature peel.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		t.Fatalf("CS-pin-old-tag: want exit 0 (must not tidy missing unpublished color); exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.DepPath == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, DepPath, and LeafModulePath required")
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "does not contain package") {
		t.Fatalf("CS-pin-old-tag: tidy must not report missing unpublished package\ncombined:\n%s", out)
	}
	if strings.Contains(lower, "unknown revision") {
		t.Fatalf("CS-pin-old-tag: unknown revision after apply\ncombined:\n%s", out)
	}
	stalePin := "pin " + labelSkills + " <- " + labelDotPkgs + " @ " + unwindApplyOldTag
	if strings.Contains(out, stalePin) {
		t.Fatalf("CS-pin-old-tag: must not pin mid @ LatestTag %s after leaf land\ncombined:\n%s",
			unwindApplyOldTag, out)
	}

	colorPath := filepath.Join(req.SecondRepo, "color", "color.go")
	if _, err := os.Stat(colorPath); err != nil {
		t.Fatalf("CS-pin-old-tag: free main missing landed color/color.go: %v", err)
	}
	assertLeafMainAdvancedAndTagged(t, req)
	assertFreeTagNextBeforeSkillsPinOfFree(t, out)
	assertMidPinnedToFreeLeaf(t, req)
	assertFeatureWIPLanded(t, req.DepPath)
	assertGoModCommittedClean(t, req.DepPath)
}
```
