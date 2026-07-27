## Expected

- Non-zero exit.
- Stderr indicates mutual exclusion / not valid / cannot combine.
- Mentions `--main` and/or `--merge-back` (generic “`--main` is mutually exclusive with other modes” is OK today).

## Errors

- `--main` cannot compose with `--merge-back`.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --main --merge-back; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "cannot") {
		t.Fatalf("stderr should indicate exclusion, got %q", se)
	}
	if !strings.Contains(se, "--main") &&
		!strings.Contains(strings.ToLower(se), "main") &&
		!strings.Contains(se, "--merge-back") &&
		!strings.Contains(se, "merge-back") {
		t.Fatalf("stderr should mention --main and/or --merge-back, got %q", se)
	}
}
```
