## Expected

- Non-zero exit (dirty cascade target).
- Stderr (and stdout) free of ANSI CSI sequences (D9 pipe default).
- Prefer `Error:` + dirty language; soft on exact wording.

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
		t.Fatalf("expected non-zero on dirty cascade; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertNoANSIEscape(t, resp.Stdout, "stdout")
	assertNoANSIEscape(t, resp.Stderr, "stderr")

	combinedLower := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if !strings.Contains(combinedLower, "uncommitted") && !strings.Contains(combinedLower, "clean") &&
		!strings.Contains(combinedLower, "dirty") {
		t.Fatalf("expected dirty language; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	assertFileExists(t, req.ExternalWtDir)
	assertFileExists(t, req.WtDir)
}
```
