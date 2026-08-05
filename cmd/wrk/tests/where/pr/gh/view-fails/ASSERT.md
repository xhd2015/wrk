## Expected

- Non-zero exit.
- Stdout empty.
- Stderr indicates `gh` / `pr view` failure (surface gh error; soft substrings:
  `gh`, `view`, `pull request`, or `failed`).

## Errors

- Failed `gh pr view` aborts location lookup.

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
		t.Fatalf("gh view failure must be non-zero; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "gh") && !strings.Contains(se, "view") &&
		!strings.Contains(se, "pull") && !strings.Contains(se, "fail") {
		t.Fatalf("stderr should surface gh view failure; got %q", resp.Stderr)
	}
}
```
