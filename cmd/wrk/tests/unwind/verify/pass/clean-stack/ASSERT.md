## Expected Output

```text
==== unwind verify ====
  dirty-peel          pass
  needs-land          pass
  owned-changed       pass   (0 modules)
  require-drift       pass   (0 edges)
  droppable-replace   pass   (0 replaces)
  cascade-pending     pass

==== status summary ====
checks: 6  pass: 6  fail: 0  warn: 0
result: pass
```

(Exact spacing implementer-owned; check ids + `pass` + `result: pass` locked.)

## Expected

- Exit code 0.
- Human banners present.
- All six catalog checks show `pass`.
- `result: pass`.
- Stdout ends with trailing `\n`.
- No ANSI without color flags.
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertVerifyHumanBanners(t, resp.Stdout)
	assertVerifyAllChecksPass(t, resp.Stdout)
	assertVerifyResult(t, resp.Stdout, "pass")
	assertVerifyStdoutTrailingNL(t, resp.Stdout)
	assertVerifyNoANSI(t, resp.Stdout)
	assertVerifyZeroMutations(t, req)
}
```
