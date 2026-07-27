
## Expected Output

```
would: git push origin main
```

## Expected

- Exit code 0.
- Stdout is the dry-run push plan line only (no human `pushed …` confirm).
- Stderr empty.
- Origin `refs/heads/main` unchanged from pre-run snapshot.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	assert.Output(t, resp.Stdout, v2StdoutTemplate(wouldPushBranchLine("origin", "main")))
	if strings.Contains(resp.Stdout, "pushed ") {
		t.Fatalf("dry-run must not print pushed confirmation; got %q", resp.Stdout)
	}

	beforeBytes, err := os.ReadFile(filepath.Join(req.WorkRoot, "origin-main-before"))
	if err != nil {
		t.Fatalf("read origin snapshot: %v", err)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/main")
	if after != before {
		t.Fatalf("origin/main mutated under --dry-run: before %s after %s", before, after)
	}
}
```
