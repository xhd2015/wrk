## Expected

- Non-zero exit.
- Stdout empty.
- Stderr requires full GitHub pull request URL (scheme required).

## Errors

- `github.com/owner/repo/pull/N` without scheme is not accepted.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("scheme-less URL must fail; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	ok := strings.Contains(se, "full github") ||
		strings.Contains(se, "pull request url") ||
		(strings.Contains(se, "github") && strings.Contains(se, "url")) ||
		strings.Contains(se, "invalid")
	if !ok {
		t.Fatalf("stderr should require full GitHub PR URL; got %q", resp.Stderr)
	}
}
```
