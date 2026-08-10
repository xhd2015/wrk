## Expected

- Exit code 0.
- Cascade succeeds despite pre-existing go.mod WIP because `--add-all` is set:
  - Shared tag `pkgs/shared/v0.0.2` exists.
  - Root require bumped to `v0.0.2`; **keep local replace**.
  - Cascade pin commit present; go.mod/go.sum committed clean afterward.
  - Root tag `v0.0.2` with commit-before-tag (pin ancestor of tag).
  - Origin pushed (main + tags).

## Side Effects

- Pin selective commit may include the WIP go.mod line when staged via `--add-all`.
- Without cascade apply, current TagNextAll path or `--add-all requires --commit`
  fails → **RED**.

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
	assertGoModCommittedClean(t, req.MainRepo)

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
