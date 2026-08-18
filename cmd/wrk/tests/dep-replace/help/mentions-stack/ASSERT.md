## Expected

- Help succeeds (exit 0 preferred).
- Combined help text mentions `--dep-replace`.
- Mentions unwind/stack (or equivalent) on the `--dep-replace` usage lines
  (not only the generic `--dry-run` partner catalog).

## Exit Code

- 0 preferred

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--dep-replace") {
		t.Fatalf("help must mention --dep-replace; exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !helpMentionsStackWithDepReplace(help) {
		t.Fatalf("help must mention unwind/stack (or equivalent) for --dep-replace; got:\n%s", help)
	}
}

func helpMentionsStackWithDepReplace(help string) bool {
	lines := strings.Split(help, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "--dep-replace") {
			continue
		}
		end := i + 3
		if end > len(lines) {
			end = len(lines)
		}
		window := strings.ToLower(strings.Join(lines[i:end], "\n"))
		if strings.Contains(window, "unwind") || strings.Contains(window, "stack") {
			return true
		}
	}
	return false
}
```
