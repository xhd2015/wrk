## Expected

- Exit code 0.
- Cascade free-first side effects (same as C-AP2 core):
  - Leaf tagged `v0.0.2` on leaf main.
  - Root consumer require `example.com/dot-pkgs` bumped to `v0.0.2`.
  - Cascade pin commit on root history.
  - Root go.mod/go.sum committed clean after pin.
- Reinstall-local tail after cascade:
  - No `unknown revision`.
  - No `go mod tidy required` / `failed to execute go mod tidy`.
  - Soft empty reinstall (no candidates) is OK; if summary printed, `failed 0`.

## Side Effects

- Same land/cascade mutations as multi-repo clean apply; plus reinstall scan of
  collected mains (may no-op installs under empty GOBIN).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("C-RI2 multi-repo + reinstall: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.SecondRepo == "" || req.MainRepo == "" {
		t.Fatal("SecondRepo and MainRepo required")
	}

	// Cascade free-first still holds with reinstall tail attached.
	assertLeafMainAdvancedAndTagged(t, req)
	assertRequireBumped(t, req)
	assertCascadePinCommitPresent(t, req.MainRepo, unwindDotPkgsModule, unwindApplyNextTag)
	assertGoModCommittedClean(t, req.MainRepo)

	if !tagRefExists(t, req.SecondRepo, unwindApplyNextTag) {
		t.Fatalf("leaf tag %s missing after cascade+reinstall", unwindApplyNextTag)
	}

	assertReinstallTailNoHardFail(t, resp)
}
```
