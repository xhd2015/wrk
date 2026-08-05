## Expected

- Non-zero exit.
- Stdout empty (must not resolve the PR path via equals binding).
- Stderr indicates `--pr` equals form is invalid / not accepted (or requires
  positional full URL). Soft: mention `--pr` or equals / invalid.

## Errors

- `--pr` is Bool; URL must be a separate positional.

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
		t.Fatalf("--pr=URL must fail; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "--pr") && !strings.Contains(se, "equals") &&
		!strings.Contains(se, "invalid") && !strings.Contains(se, "url") {
		t.Fatalf("stderr should reject --pr equals form; got %q", resp.Stderr)
	}
}
```
