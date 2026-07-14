## Expected

- Exit code 0; stderr empty (or only non-fatal noise; prefer empty).
- Composition is not rejected as mutually exclusive / invalid flag combo.
- Non-empty status stdout; may show `Remote: (no upstream)`.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	// --fetch must not be rejected as exclusive with the --main --status pair.
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("--fetch should be allowed with --main --status; stderr=%q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "only valid with") {
		t.Fatalf("--fetch should be valid with --status (even with --main); stderr=%q", resp.Stderr)
	}
	if resp.Stdout == "" {
		t.Fatal("expected non-empty status stdout")
	}
}
```
