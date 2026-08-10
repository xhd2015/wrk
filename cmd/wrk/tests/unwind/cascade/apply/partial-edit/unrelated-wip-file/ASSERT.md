## Expected

- Exit code 0.
- Cascade pin commit present; **name-only files** are only `go.mod` / `go.sum`.
- Pin commit go.mod has require bump and **no** WIP marker.
- WT go.mod: WIP marker + surgical require bump + keep replace (still dirty).
- Unrelated `WIP_NOTES.md` still present and uncommitted (not in HEAD).

## Side Effects

- Partial edit does not swallow non-module WIP into the cascade pin commit.

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
		t.Fatalf("partial-edit with unrelated WIP: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	assertCascadePinCommitPresent(t, req.MainRepo, cascadeSharedModule, unwindApplyNextTag)
	assertPinCommitFilesOnlyModSum(t, req.MainRepo)
	assertPinCommitBaseNoWIP(t, req)
	assertPartialEditWTPreserved(t, req)
	assertUnrelatedWIPStillPresent(t, req)

	if !tagRefExists(t, req.MainRepo, cascadeSharedNextTag) {
		t.Fatalf("missing shared tag %s", cascadeSharedNextTag)
	}
}
```
