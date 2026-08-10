## Expected

- Exit code **non-zero**.
- Failure is on the **partial-edit pin/tidy path** (not the obsolete P2 hard
  Error that refuses dirty go.mod without `--add-all` before attempting pin).
  Combined output should mention tidy / go mod / revision / proxy-style
  diagnostics typical of `go mod tidy` failure after pin.
- Consumer **WT go.mod** restored **byte-identical** to pre-run WIP snapshot
  (marker present; require still at old version; no half-applied Base pin).

## Side Effects

- Peel/tag of free leaf may have succeeded before consumer tidy fails — OK
  (and expected evidence that cascade ran past free-module tag-next).
- Must **not** leave worktree go.mod as Base-only (WIP wiped) or Base+pin
  without restore.
- **Classic TDD RED** while product still hard-fails dirty Base before pin
  (P2 message about uncommitted go.mod / `--add-all`).

## Errors

- Mid-partial tidy failure → restore + non-zero (not dirty-Base hard refuse).

## Exit Code

- non-zero

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("tidy-fail mid partial-edit must non-zero; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	// P2 hard refuse of dirty Base must be gone once partial edit exists.
	// Today product still errors with uncommitted + --add-all → RED (correct TDD).
	dirtyHardFail := (strings.Contains(combined, "uncommitted") || strings.Contains(combined, "has uncommitted changes")) &&
		(strings.Contains(combined, "add-all") || strings.Contains(combined, "--add-all"))
	if dirtyHardFail {
		t.Fatalf("partial-edit tidy-fail path not reached: still hard-failing dirty go.mod without --add-all (P2)\nstderr=%q\nstdout=%q",
			resp.Stderr, resp.Stdout)
	}

	// Prefer tidy / pin diagnostic once partial edit attempts Base pin.
	hasTidy := strings.Contains(combined, "tidy") || strings.Contains(combined, "go mod")
	hasGoDiag := strings.Contains(combined, "unknown revision") ||
		strings.Contains(combined, "no such file") ||
		strings.Contains(combined, ".zip") ||
		strings.Contains(combined, "reading ") ||
		strings.Contains(combined, "not found") ||
		strings.Contains(combined, "invalid version")
	if !hasTidy && !hasGoDiag {
		t.Fatalf("expected tidy/pin failure diagnostic after partial-edit Base pin\nstderr=%q\nstdout=%q",
			resp.Stderr, resp.Stdout)
	}

	// Core contract: exact WIP restore (no half-mutated WT).
	assertPartialEditWIPRestoredExact(t, req)

	goMod := filepath.Join(req.MainRepo, "go.mod")
	content := readFile(t, goMod)
	if !strings.Contains(content, cascadeGoModWIPMarker) {
		t.Fatalf("restored go.mod must still contain WIP marker\n%s", content)
	}
	if req.LeafModulePath != "" && req.OldRequireVersion != "" {
		got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
		if got != req.OldRequireVersion {
			t.Fatalf("restored require %s = %q, want pre-run %s\ngo.mod:\n%s",
				req.LeafModulePath, got, req.OldRequireVersion, content)
		}
	}
}
```
