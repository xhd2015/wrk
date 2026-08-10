## Expected Output

```text
==== unwind verify ====
  dirty-peel          FAIL
  …
==== status summary ====
…
result: fail
```

## Expected

- Exit code **1** (non-zero).
- Human banners present.
- `dirty-peel` shows `FAIL` (uppercase).
- `result: fail`.
- No `Error:` on stdout/stderr for logical FAIL.
- Stdout trailing `\n`.
- Zero mutations; DIRTY still present.

## Side Effects

- None.

## Exit Code

- 1

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	if resp.ExitCode != 1 {
		t.Fatalf("check FAIL must exit 1; got %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertVerifyHumanBanners(t, resp.Stdout)
	assertVerifyCheckStatus(t, resp.Stdout, "dirty-peel", "FAIL")
	assertVerifyResult(t, resp.Stdout, "fail")
	assertVerifyNoLogicalErrorPrefix(t, resp)
	assertVerifyStdoutTrailingNL(t, resp.Stdout)
	assertVerifyZeroMutations(t, req)
	assertFileExists(t, filepath.Join(req.MainRepo, "DIRTY"))
}
```
