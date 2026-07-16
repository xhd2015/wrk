## Expected

- Non-zero exit.
- Stderr mentions `--json`.
- Stderr indicates `--json` is not valid without `--tag-next` (or not valid with bare push).
- Must not push or print success confirm line.

## Errors

- `--json` remains host-bound to bare `--tag-next`.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --push --json, got 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--json") {
		t.Fatalf("stderr should mention --json, got %q", se)
	}
	if strings.Contains(resp.Stdout, "pushed main → origin/main") {
		t.Fatalf("must not succeed-push under --json reject; stdout=%q", resp.Stdout)
	}
}
```
