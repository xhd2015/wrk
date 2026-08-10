## Expected

- Exit code 0.
- Tags: `pkgs/shared/v0.0.2` and `pkgs/other/v0.0.2` exist.
- WT go.mod: WIP marker preserved; **both** `example.com/root/shared` and
  `example.com/root/other` requires at `v0.0.2`; keep both local replaces;
  go.mod still dirty.
- Cascade pin commit history records both require bumps (one or more pin
  commits); pin trees never include the WIP marker.
- Root next tag `v0.0.2` after pins (commit-before-tag).

## Side Effects

- Sequential partial-edit pins must not clobber earlier surgical bumps when
  restoring WIP between steps / at end.

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
		t.Fatalf("sequential two pins partial-edit: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	assertSequentialPinsOnWTAndCommits(t, req)

	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("missing root tag %s after sequential cascade", unwindApplyNextTag)
	}
	assertCommitBeforeTag(t, req.MainRepo, unwindApplyNextTag)

	if req.OriginBare != "" {
		assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	}
}
```
