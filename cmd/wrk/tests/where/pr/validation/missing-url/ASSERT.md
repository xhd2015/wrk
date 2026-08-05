## Expected

- Non-zero exit.
- Stdout empty.
- Stderr indicates a full GitHub pull request URL is required (prefer substring
  `full GitHub` or `pull request URL` / `GitHub` + `URL`).

## Errors

- Exactly one positional (full PR URL) is required for `--where --pr`.

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
		t.Fatalf("expected non-zero when URL missing; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	ok := strings.Contains(se, "full github") ||
		strings.Contains(se, "pull request url") ||
		(strings.Contains(se, "github") && strings.Contains(se, "url")) ||
		(strings.Contains(se, "requires") && strings.Contains(se, "url"))
	if !ok {
		t.Fatalf("stderr should require full GitHub PR URL; got %q", resp.Stderr)
	}
}
```
