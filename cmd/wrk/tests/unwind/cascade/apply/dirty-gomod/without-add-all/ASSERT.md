## Expected

- Exit code **0** (partial edit succeeds for ordinary WIP).
- Free-first cascade side effects:
  - Shared path-scope tag `pkgs/shared/v0.0.2` exists.
  - Cascade pin commit present: subject prefix `wrk: cascade pin `.
  - **Pin commit go.mod** = Base + pin + tidy: require bumped; **no** WIP marker.
  - **Worktree go.mod** still has WIP marker + surgical require bump to next +
    keep local replace; go.mod remains dirty (uncommitted WIP).
  - Root tag `v0.0.2` with pin commit as ancestor; origin push OK.

## Side Effects

- Partial edit: save WIP → write Base → pin+tidy → selective commit → restore WIP
  with surgical require bumps only (no tidy on restored WT).
- **Not** hard Error (P2 behavior obsolete for normal WIP — D11).
- **Classic TDD RED** until product implements partial edit (today hard-fails).

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
		t.Fatalf("partial-edit dirty go.mod without --add-all must succeed; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	if !tagRefExists(t, req.MainRepo, cascadeSharedNextTag) {
		t.Fatalf("missing shared tag %s\nstderr=%q stdout=%q",
			cascadeSharedNextTag, resp.Stderr, resp.Stdout)
	}

	assertCascadePinCommitPresent(t, req.MainRepo, cascadeSharedModule, unwindApplyNextTag)
	assertPinCommitBaseNoWIP(t, req)
	assertPartialEditWTPreserved(t, req)

	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("missing root tag %s after partial-edit cascade", unwindApplyNextTag)
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
