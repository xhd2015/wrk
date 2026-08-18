## Expected

- Help succeeds (exit 0 preferred).
- Combined help text mentions `--dep-update`.
- Mentions unwind/stack (or equivalent stack inventory wording) for `--dep-update`.
- Mentions `--dry-run`.

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
	if !strings.Contains(help, "--dep-update") {
		t.Fatalf("help must mention --dep-update; exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(help, "--dry-run") {
		t.Fatalf("help must mention --dry-run; exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !helpMentionsStackWithDepUpdate(help) {
		t.Fatalf("help must mention unwind/stack (or equivalent) for --dep-update; got:\n%s", help)
	}
}

func helpMentionsStackWithDepUpdate(help string) bool {
	lines := strings.Split(help, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		// Usage / description of --dep-update itself, not the generic
		// --dry-run partner catalog that lists every mode.
		if !strings.HasPrefix(trim, "--dep-update") {
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
