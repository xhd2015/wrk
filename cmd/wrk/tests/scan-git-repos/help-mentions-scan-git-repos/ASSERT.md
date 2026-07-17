## Expected

- Exit code 0.
- Help text (stdout and/or stderr) contains `--scan-git-repos`.
- Help text also contains `--no-cache`.
- Help documents that bare `--scan-git-repos` defaults to home (`~`) when no ROOT is given
  (soft match common phrasings; existing `~/.wrk` alone is not enough).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--scan-git-repos") {
		t.Fatalf("help must mention --scan-git-repos; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(help, "--no-cache") {
		t.Fatalf("help must mention --no-cache; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, help, `<contains>
--scan-git-repos
</contains>`)
	// P1: default scan root is ~ (home), not ~/Projects — document on help line.
	// Soft forms so implementers can choose wording without false greens from ~/.wrk alone.
	if !scanHelpMentionsDefaultHome(help) {
		t.Fatalf("help for --scan-git-repos must document default root ~ (home); got %q", help)
	}
}

// scanHelpMentionsDefaultHome is true when the --scan-git-repos help line itself
// documents default root ~ / home. Whole-help search is too loose (WRK_HOME is
// "default: ~/.wrk").
func scanHelpMentionsDefaultHome(help string) bool {
	for _, line := range strings.Split(help, "\n") {
		if !strings.Contains(line, "--scan-git-repos") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(line, "~") ||
			strings.Contains(line, "$HOME") ||
			strings.Contains(lower, "home") {
			return true
		}
	}
	return false
}
```
