## Expected

- Exit code 0.
- Help text contains `--bring`.
- Soft: help near `--bring` mentions multi / multiple / repeatable / more than one / dependencies (plural) / can be repeated.

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
	if !multiBringHelpMentionsRepeatable(help) {
		t.Fatalf("help for --bring should mention multi/repeatable form; got %q", help)
	}
}

func multiBringHelpMentionsRepeatable(help string) bool {
	for _, line := range strings.Split(help, "\n") {
		if !strings.Contains(line, "--bring") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "multiple") ||
			strings.Contains(lower, "repeat") ||
			strings.Contains(lower, "more than one") ||
			strings.Contains(lower, "several") ||
			strings.Contains(lower, "dependencies") ||
			strings.Contains(lower, "can be") ||
			// e.g. --bring path  (repeatable)
			strings.Contains(lower, "repeatable") {
			return true
		}
	}
	// Whole-help soft fallback when description spans lines near --bring.
	lower := strings.ToLower(help)
	idx := strings.Index(lower, "--bring")
	if idx < 0 {
		return false
	}
	// Window after first --bring mention.
	window := lower[idx:]
	if len(window) > 400 {
		window = window[:400]
	}
	return strings.Contains(window, "multiple") ||
		strings.Contains(window, "repeat") ||
		strings.Contains(window, "dependencies") ||
		strings.Contains(window, "more than one")
}
```
