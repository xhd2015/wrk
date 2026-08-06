## Expected

- Exit code ≠ 0 (tidy fails after pin requires next version missing from modproxy).
- Combined stdout/stderr:
  - still mentions `go mod tidy` (and typically the consumer module path)
  - **contains a concrete go diagnostic** such as `no such file`, `unknown
    revision`, `reading …`, `.zip` / `.info`, or similar non-empty go body
  - **must not** be only `failed to execute go mod tidy: exit status 1` with no
    child diagnostic

## Side Effects

- Peel may have landed/tagged leaf before tidy fails (fail-fast at pin/tidy).
- No requirement that consumer go.mod is left clean after partial pin.

## Exit Code

- non-zero

## Errors

- Tidy-phase hard error with go child stderr surfaced.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	// Classic TDD RED: current product returns tidy failure as
	// "failed to execute go mod tidy: exit status 1" without go child stderr.
	assertTidyErrorSurfacesGoStderr(t, resp)
}
```
