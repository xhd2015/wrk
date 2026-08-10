## Expected

- Exit code 0.
- Free-first cascade side effects on the monorepo:
  1. Path-scope tag `pkgs/shared/v0.0.2` exists (shared leaf tagged).
  2. Nested **tools** `go.mod` requires `example.com/root/shared` at `v0.0.2`.
  3. **Keep local replace** on tools (`replace … => ../pkgs/shared` still present).
  4. **No** cascade `tag-next` for `example.com/root/tools` (skip consumer).
  5. Root `go.mod` does **not** gain a shared require (tools owns the edge).
  6. Cascade pin commit on history mentioning shared / pin version.
  7. tools/go.mod (and go.sum if present) committed clean after pin.
- Reinstall-local tail:
  1. No `unknown revision`.
  2. No `go mod tidy required` / `failed to execute go mod tidy` / updates-needed.
  3. Seeded `tools/cmd/tool` reinstalls with summary `failed 0` (and N≥1 when summary present).
- Optional apply vocabulary (C-RI3 light): stdout shows cascade pin / tag-next for shared before reinstall work.

## Side Effects

- Cascade mutates main: shared tag, tools pin commit + require bump; does **not**
  drop tools local replace; does **not** tag tools.
- Reinstall may write under isolated `GOBIN={WorkRoot}/gobin`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("C-RI1 nested-skip + reinstall: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	// 1) Free-first: shared path-scope tag exists.
	if !tagRefExists(t, req.MainRepo, cascadeSharedNextTag) {
		t.Fatalf("missing free-first shared tag %s on main\nstdout:\n%s\nstderr:\n%s",
			cascadeSharedNextTag, resp.Stdout, resp.Stderr)
	}

	// 2–5) Nested skip consumer pin + keep-replace; root not forced into edge.
	assertNestedSkipConsumerPinned(t, req)
	assertNoCascadeTagNextForModule(t, resp.Stdout, cascadeToolsModule)

	// 6) Cascade pin commit present.
	assertCascadePinCommitPresent(t, req.MainRepo, cascadeSharedModule, unwindApplyNextTag)

	// 7) tools module go.mod/go.sum clean after pin commit.
	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain", "--",
		filepath.Join(cascadeToolsDir, "go.mod"),
		filepath.Join(cascadeToolsDir, "go.sum"))
	if strings.TrimSpace(status) != "" {
		t.Fatalf("tools go.mod/go.sum must be committed after cascade pin; porcelain:\n%s", status)
	}

	// Reinstall tail: original failure mode must stay fixed.
	assertReinstallInstalledAtLeastOne(t, resp)
	assertReinstallTailNoHardFail(t, resp)

	// C-RI3 light: cascade apply vocabulary for shared present (not dry-run would:).
	out := resp.Stdout
	if !strings.Contains(out, "tag-next "+cascadeSharedModule) &&
		!strings.Contains(out, cascadeSharedNextTag) {
		// Tag creation is authoritative via tagRefExists; soft-check apply banner.
		_ = cascadeSharedModule
	}
	if !strings.Contains(out, "pin ") && !hasCascadePin(out, cascadeToolsModule, cascadeSharedModule) {
		// Apply pin log uses basename labels; accept either form.
		if !strings.Contains(strings.ToLower(out), "pin ") {
			t.Fatalf("apply cascade should log pin for nested consumer\nstdout:\n%s", out)
		}
	}
}
```
