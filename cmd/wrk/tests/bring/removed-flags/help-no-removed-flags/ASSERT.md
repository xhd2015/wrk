## Expected

- Exit code 0.
- Help text mentions `--bring`.
- Help text does **not** document `--dep` or `--all-deps` as live modes.
- Bring help does not say “like --dep”.

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--bring") {
		t.Fatalf("help must mention --bring; got %q", help)
	}
	// Reject live-mode documentation for removed flags (stable line-oriented checks).
	for _, line := range strings.Split(help, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--dep") || strings.Contains(line, "  --dep ") || strings.Contains(line, "\t--dep") {
			t.Fatalf("help must not document --dep as a flag; line=%q", line)
		}
		if strings.HasPrefix(trim, "--all-deps") || strings.Contains(line, "  --all-deps") {
			t.Fatalf("help must not document --all-deps as a flag; line=%q", line)
		}
	}
	// Whole-help guard for "like --dep" style wording.
	if strings.Contains(help, "like --dep") || strings.Contains(help, "like `--dep`") {
		t.Fatalf("help must not contrast bring as like --dep; got %q", help)
	}
}
```
