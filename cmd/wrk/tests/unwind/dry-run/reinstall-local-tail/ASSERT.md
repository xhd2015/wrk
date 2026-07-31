## Expected Output

The ordinary dry-run peel plan ends with a newline:

```
==== unwind (dry-run) ====
would: peel root
```

## Expected

- Exit code 0.
- The ordinary single-repository peel plan contains `would: peel root`.
- `--reinstall-local` is accepted with `--unwind`; diagnostics do not describe it as mutually exclusive.
- Zero mutations: HEAD is unchanged and the dirty marker remains.

## Side Effects

- None. This is a dry run; no reinstall executes.

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
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if strings.Contains(combined, "mutually exclusive") {
		t.Fatalf("unwind+reinstall-local must be accepted; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
