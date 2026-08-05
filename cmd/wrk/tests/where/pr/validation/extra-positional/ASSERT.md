## Expected

- Non-zero exit.
- Stdout empty.
- Stderr mentions unexpected arguments (preferred exact family:
  `wrk: unexpected arguments`).

## Errors

- Exactly one positional allowed for `--where --pr`.

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
		t.Fatalf("extra positional must fail; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "unexpected") && !strings.Contains(se, "argument") {
		t.Fatalf("stderr should mention unexpected arguments; got %q", resp.Stderr)
	}
}
```
