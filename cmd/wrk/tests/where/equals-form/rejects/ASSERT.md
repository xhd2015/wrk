## Expected

- Non-zero exit.
- Stdout empty (must **not** print the saved project path — equals form is not basename lookup).
- Stderr indicates failure involving `--where` / equals or invalid value form (soft wording).

## Errors

- `--where=spl` is invalid after Bool + positional binding.

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
		t.Fatalf("expected non-zero exit for --where=spl, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		// Reject false-GREEN: String binding would print the saved path.
		t.Fatalf("stdout should be empty (equals form must not lookup basename), got %q", resp.Stdout)
	}
	wantPath := resolvePath(t, req.MainRepo)
	if strings.Contains(resp.Stdout, wantPath) {
		t.Fatalf("stdout must not contain saved path %q under equals form; got %q", wantPath, resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "--where") {
		t.Fatalf("stderr should mention --where, got %q", resp.Stderr)
	}
}
```
