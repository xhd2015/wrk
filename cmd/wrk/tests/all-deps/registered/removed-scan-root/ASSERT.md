## Expected

- Non-zero exit code.
- Stderr mentions `--scan-root` (unknown/unexpected flag or flag removed).

## Exit Code

- Non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "scan-root") {
		t.Fatalf("stderr should mention scan-root, got %q", resp.Stderr)
	}
}
```