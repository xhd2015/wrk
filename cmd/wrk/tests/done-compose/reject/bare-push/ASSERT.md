## Expected

- Non-zero exit (regression: bare `--push` stays invalid).
- Stderr mentions `--push` and a validity/host policy (e.g. only valid with `--tag-next` and/or primary modes).
- Must not create worktrees or tags as a silent no-args create.

## Errors

- `--push` alone is not a valid mode.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for bare --push, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--push") {
		t.Fatalf("stderr should mention --push, got %q", se)
	}
	if !strings.Contains(se, "only valid") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "requires") &&
		!strings.Contains(se, "mutually exclusive") {
		t.Fatalf("stderr should indicate --push is not standalone, got %q", se)
	}
}
```
