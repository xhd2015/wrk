## Expected

- Exit code 0.
- Cascade pin (agent-pro shape: nested cmd → parent, keep `replace => ..`):
  1. Free checkout **cmd/** `go.mod` requires free parent at `v0.0.2`.
  2. **Keep local replace** (`=> ..`) still present.
  3. **No** cascade `tag-next` for `example.com/dot-pkgs/cmd-harness`.
  4. Cascade pin commit present for free parent @ pin version (on free history).
- Reinstall-local tail (scans free **main** via useMain):
  1. No `unknown revision`.
  2. No `go mod tidy required` / `failed to execute go mod tidy` / `updates to go.mod needed`.
  3. Seeded `cmd/tool` reinstalls with summary `failed 0` (N≥1 when summary present).

## Side Effects

- Pin may commit on free linked worktree Path; reinstall builds free main modules.
- Product must leave free main installable (pin+tidy visible to reinstall scan root)
  or reinstall must use the checkout that was pinned/tidied.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("C-RI3 nested-cmd-requires-parent + reinstall: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.SecondRepo == "" {
		t.Fatal("SecondRepo (free main) required")
	}

	// 1–3) Nested cmd pin parent require + keep replace => .. on free **main**
	// (clean linked Path pins MainRepo so reinstall useMain sees it).
	assertNestedCmdRequiresParentPinned(t, req)
	assertNoCascadeTagNextForModule(t, resp.Stdout, cascadeCmdHarnessModule)

	// 4) Cascade pin commit on free main history.
	assertCascadePinCommitPresent(t, req.SecondRepo, unwindDotPkgsModule, unwindApplyNextTag)

	// Reinstall tail: must install nested cmd package without tidy/sum diagnostics.
	assertReinstallInstalledAtLeastOne(t, resp)
	assertReinstallTailNoHardFail(t, resp)

	// Explicit go.mod/go.sum hygiene failures (agent-pro class).
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if strings.Contains(combined, "updates to go.mod needed") ||
		strings.Contains(combined, "missing go.sum entry") {
		t.Fatalf("reinstall must not hit go.mod/go.sum hygiene failure after cascade pin of nested cmd←parent\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
}
```
