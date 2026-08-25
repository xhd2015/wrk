## Expected

- Exit code 0.
- Stdout is **root** usage (wrk — git worktree helper).
- Mentions `--bash-integration` and dedicated help (`--bash-integration --help` or equivalent).
- Stderr is empty.
- Stdout ends with trailing `\n`.

## Side Effects

- Read-only (`--help` only).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --help, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("stderr should be empty for --help, got %q", resp.Stderr)
	}
	assertStdoutEndsWithNewline(t, resp.Stdout)
	lower := strings.ToLower(resp.Stdout)
	if !strings.Contains(lower, "--bash-integration") {
		t.Fatalf("root help must mention --bash-integration:\n%s", resp.Stdout)
	}
	if !strings.Contains(lower, "bash integration") {
		t.Fatalf("root help must mention Bash integration section:\n%s", resp.Stdout)
	}
	if !strings.Contains(lower, "--bash-integration --help") &&
		!strings.Contains(lower, "--bash-integration -h") {
		t.Fatalf("root help should point to bash-integration dedicated help:\n%s", resp.Stdout)
	}
}
```
