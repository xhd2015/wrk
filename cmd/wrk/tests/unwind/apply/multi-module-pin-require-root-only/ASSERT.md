## Expected

- Exit code 0.
- Dep main advanced with root feature content; local root tag `v0.0.2` at dep main HEAD;
  bare origin has main + root tag.
- **Pin log:** exactly **one** `pin root <- dep` line (required module pairing only).
  Cartesian pin of root+nested prints the same basename line twice — primary RED surface
  (tidy may drop an unused force-added nested require, so go.mod alone can mask the bug).
- Consumer Path go.mod (`req.MainRepo`, primary is main):
  - requires `example.com/dep` at `v0.0.2`
  - **must not** require `example.com/dep/nested` at any version
  - no replace for nested (or root)
- `go mod tidy` succeeded (implied by exit 0 + offline offline proxy for root next).

## Side Effects

- Peel lands nested dep WT; tag-next creates root next tag (nested scope may stay
  at `nested/v0.0.1` when only root content changed).
- Pin edits only the required dep module path; must not invent nested require from
  root tag version; must not pin log spam for non-required modules.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		// Possible today if nested force-add is kept and tidy cannot resolve next.
		t.Fatalf("multi-module require-root-only: want exit 0 (no force-add nested pin); exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" || req.LeafModulePath == "" || req.NestedModulePath == "" {
		t.Fatal("MainRepo, LeafModulePath, NestedModulePath required")
	}
	if resolvePath(t, req.RepoDir) != resolvePath(t, req.MainRepo) {
		t.Fatalf("expect primary Path == MainRepo; RepoDir=%s MainRepo=%s",
			req.RepoDir, req.MainRepo)
	}
	// Primary RED: cartesian pin prints pin root <- dep once per dep module dir.
	assertExactlyOnePinLine(t, out, labelRoot, labelDep)
	// Dep ship still OK (root package advanced + root tag).
	assertLeafMainAdvancedAndTagged(t, req)
	// go.mod: root bumped; nested must stay absent (if still present after tidy).
	assertConsumerPinnedRootOnly(t, req)
}
```
