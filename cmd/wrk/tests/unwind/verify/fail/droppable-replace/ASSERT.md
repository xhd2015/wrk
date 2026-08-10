## Expected

- Exit code non-zero (prefer 1).
- Human banners present.
- `droppable-replace` shows `FAIL`.
- Intra-repo-style keep-local replaces are **not** the subject of this leaf
  (external stack replace only).
- `result: fail`.
- No `Error:` for logical FAIL.
- Zero mutations (replace still present in go.mod).

## Side Effects

- None.

## Exit Code

- 1

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertVerifyHumanBanners(t, resp.Stdout)
	assertVerifyCheckStatus(t, resp.Stdout, "droppable-replace", "FAIL")
	assertVerifyResult(t, resp.Stdout, "fail")
	assertVerifyNoLogicalErrorPrefix(t, resp)
	assertVerifyZeroMutations(t, req)
}
```
