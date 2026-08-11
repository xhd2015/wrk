## Expected

- Help succeeds (exit 0 preferred).
- Combined help text mentions `--all`.
- Mentions of `--all` appear in context of `--dep-update` (same usage block /
  nearby lines), not only an unrelated flag.

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
	if !strings.Contains(help, "--all") {
		t.Fatalf("help must mention --all (partner of --dep-update); exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	// Require co-occurrence: some line (or a short window) ties --all to dep-update.
	if !helpMentionsAllWithDepUpdate(help) {
		t.Fatalf("help must mention --all in context of --dep-update; got:\n%s", help)
	}
}

func helpMentionsAllWithDepUpdate(help string) bool {
	lines := strings.Split(help, "\n")
	for i, line := range lines {
		if !strings.Contains(line, "--dep-update") && !strings.Contains(line, "--all") {
			continue
		}
		// Window: current line ± 2 for multi-line usage blocks.
		start := i - 2
		if start < 0 {
			start = 0
		}
		end := i + 3
		if end > len(lines) {
			end = len(lines)
		}
		window := strings.Join(lines[start:end], "\n")
		if strings.Contains(window, "--dep-update") && strings.Contains(window, "--all") {
			return true
		}
	}
	return false
}
```
