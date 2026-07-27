## Expected

- Non-zero exit (D7: preflight applies in dry-run).
- Stderr includes `Error:` and dirty/uncommitted language.
- Must **not** present a successful cascade plan: either no `would: cascade merge-back`
  line, or if headers print, still non-zero with Error (never exit 0 with would:).
- Today dry-run short-circuits before dirty check and prints `would: cascade merge-back`
  with exit 0 → **RED**.
- External + consumer still present (zero mutations).

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
		t.Fatalf("dirty cascade dry-run must fail preflight (non-zero); stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "Error:") {
		t.Fatalf("dry-run preflight failure must include structured %q; stderr:\n%s\nstdout:\n%s",
			"Error:", stderr, resp.Stdout)
	}
	combinedLower := strings.ToLower(stderr + "\n" + resp.Stdout)
	if !strings.Contains(combinedLower, "uncommitted") && !strings.Contains(combinedLower, "clean") &&
		!strings.Contains(combinedLower, "dirty") {
		t.Fatalf("expected dirty language; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}

	// Do not allow a false success plan as the only outcome: if would: cascade appears
	// it must not be the sole user-facing success path without Error: (already required).
	// Soft: prefer absence of would: cascade when blocked; Error: + non-zero is enough.
	_ = strings.Contains(resp.Stdout, "would: cascade")

	assertNoANSIEscape(t, resp.Stdout, "stdout")
	assertNoANSIEscape(t, resp.Stderr, "stderr")

	assertFileExists(t, req.ExternalWtDir)
	assertGitFileIsWorktreeLink(t, req.ExternalWtDir)
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
}
```
