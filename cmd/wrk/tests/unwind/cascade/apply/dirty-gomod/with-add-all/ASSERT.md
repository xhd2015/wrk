## Expected

- Exit code 0.
- Cascade succeeds despite pre-existing go.mod WIP when `--add-all` is set:
  - Shared tag `pkgs/shared/v0.0.2` exists.
  - Root require bumped to `v0.0.2`; **keep local replace**.
  - Cascade pin commit present; pin tree is Base+pin (**no** WIP marker).
  - WT go.mod still has WIP marker + surgical require bump (partial-edit
    isolation — `--add-all` does **not** scoop WIP into the pin commit).
  - Root tag `v0.0.2` with commit-before-tag (pin ancestor of tag).
  - Origin pushed (main + tags).

## Side Effects

- **F1 contract (2026-08-11):** `--add-all` no longer disables cascade partial-edit.
  Pin commits stay selective Base-only; WIP remains uncommitted on the WT
  (same isolation as without `--add-all`). `--add-all` still matters for
  feature gen-commit staging on peel paths, not for pin scooping.
- Historical note: older C-AP6 expected go.mod clean after pin because `--add-all`
  used to pin-on-dirty and swallow WIP; that product behavior was a bug surface
  for pin-only consumer diverge and is intentionally reversed.

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
		t.Fatalf("dirty go.mod with --add-all must succeed; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	if !tagRefExists(t, req.MainRepo, cascadeSharedNextTag) {
		t.Fatalf("missing shared tag %s\nstderr=%q stdout=%q",
			cascadeSharedNextTag, resp.Stderr, resp.Stdout)
	}
	assertRequireBumpedKeepReplace(t, req)
	assertCascadePinCommitPresent(t, req.MainRepo, cascadeSharedModule, unwindApplyNextTag)
	// F1: pin commit is Base+pin (no WIP); WT keeps WIP + surgical bump.
	assertPinCommitBaseNoWIP(t, req)
	assertPartialEditWTPreserved(t, req)

	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("missing root tag %s after cascade with --add-all", unwindApplyNextTag)
	}
	assertCommitBeforeTag(t, req.MainRepo, unwindApplyNextTag)

	if req.OriginBare == "" {
		t.Fatal("OriginBare required")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, cascadeSharedNextTag) {
		t.Fatalf("origin missing shared tag %s", cascadeSharedNextTag)
	}
	if !remoteTagExists(t, req.OriginBare, unwindApplyNextTag) {
		t.Fatalf("origin missing root tag %s", unwindApplyNextTag)
	}
}
```
