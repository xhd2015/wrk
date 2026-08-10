## Expected Output

Stdout: verify banners + all checks pass + `result: pass`.

Stderr: contains `warning:` (missing / non-git replace target).

## Expected

- Exit code 0 (do not fail verify solely for missing replace target).
- Stderr contains `warning:`.
- Human verify banners present; all catalog checks `pass`.
- Zero mutations.

## Side Effects

- None.

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
	assertExitZero(t, resp)
	assertVerifyHumanBanners(t, resp.Stdout)
	assertVerifyAllChecksPass(t, resp.Stdout)
	assertVerifyResult(t, resp.Stdout, "pass")
	if !strings.Contains(resp.Stderr, "warning:") && !strings.Contains(resp.Stderr, "Warning:") {
		t.Fatalf("missing replace target must emit warning: on stderr; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
	}
	assertVerifyZeroMutations(t, req)
}
```
