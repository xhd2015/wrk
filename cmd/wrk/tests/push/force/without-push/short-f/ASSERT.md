## Expected

- Non-zero exit.
- Stderr contains exactly the locked phrase:
  `wrk: -f/--force is only valid with --push`
- Stdout has no success push confirm line.

## Errors

- Force without `--push` is invalid (flag layer).

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for -f without --push; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, forceWithoutPushErr) {
		t.Fatalf("stderr should contain %q, got %q", forceWithoutPushErr, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "pushed ") {
		t.Fatalf("must not print push confirm without --push; stdout=%q", resp.Stdout)
	}
}
```
