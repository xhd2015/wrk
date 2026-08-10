## Expected

- Exit code **0** (cascade completes; no `tag already exists`).
- Root release tag `v0.0.2` exists on main.
- Nested tool `script/browser-debug/go.mod`:
  - require `example.com/root` at `v0.0.2`
  - keep local replace `=> ../../`
- Cascade pin commit present for dep `example.com/root` @ `v0.0.2`.
- **No** apply `tag-next` line for module `browser-debug` (pin-only consumer).
- Stderr/stdout must **not** contain `tag already exists` / `cascade tag v0.0.2`.
- Origin has main + root tag when `--push` succeeds.

## Side Effects

- Main: one new root tag + one pin commit on tool go.mod; tool module never
  receives its own release tag under root-only tagscope.
- Regression: pre-fix exited ≠ 0 with `fatal: tag 'v0.0.2' already exists` after pin.

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
	combined := resp.Stdout + "\n" + resp.Stderr
	if strings.Contains(combined, "tag already exists") ||
		strings.Contains(combined, "cascade tag "+unwindApplyNextTag) {
		t.Fatalf("root-only nested tool pin: cascade must not collide on tag %s\nstdout:\n%s\nstderr:\n%s",
			unwindApplyNextTag, resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("root-only nested tool pin: want exit 0 (pin-only nested tool); exit=%d\nstdout:\n%s\nstderr:\n%s",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	// Root free tag exists (bare next tag).
	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("missing root tag %s after cascade\nlog:\n%s",
			unwindApplyNextTag, gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-20"))
	}

	// Nested tool: pin-only — require bumped, keep-local replace.
	toolMod := filepath.Join(req.MainRepo, filepath.FromSlash(cascadeToolRelDir), "go.mod")
	got := requireVersionInGoMod(t, toolMod, unwindRootModule)
	if got != unwindApplyNextTag {
		t.Fatalf("tool require %s = %q, want %s\ngo.mod:\n%s",
			unwindRootModule, got, unwindApplyNextTag, readFile(t, toolMod))
	}
	if !goModHasReplace(t, toolMod, unwindRootModule) {
		t.Fatalf("tool go.mod must KEEP local replace for %s:\n%s",
			unwindRootModule, readFile(t, toolMod))
	}
	assertCascadePinCommitPresent(t, req.MainRepo, unwindRootModule, unwindApplyNextTag)

	// Nested tool must never be tagged (apply log form).
	assertNoCascadeTagNextForModule(t, resp.Stdout, cascadeToolModule)

	// Push tail when origin configured.
	if req.OriginBare == "" {
		t.Fatal("OriginBare required")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, unwindApplyNextTag) {
		t.Fatalf("origin missing root tag %s after --push", unwindApplyNextTag)
	}
}
```
