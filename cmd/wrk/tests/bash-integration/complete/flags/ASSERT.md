## Expected

- Exit code 0.
- Stdout contains multiple flag candidates; each line starts with `-`.
- Stdout includes key flags from usage: `--list`, `-l`, `--status`, `--bring`, `--where`, `--add`, `--rm`, `--bash-integration`.
- Stdout includes exact completion candidates `--pr`, `--title`, and `--comment` (one flag per line; not a substring of another flag).

## Side Effects

- Read-only completion.

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
	assertCompleteExitOK(t, resp)
	assertAllLinesAreFlags(t, resp.Stdout)
	assertFlagsInclude(t, resp.Stdout,
		"--list",
		"-l",
		"--status",
		"--bring",
		"--where",
		"--add",
		"--rm",
		"--bash-integration",
		"--here",
	)
	// P3: exact line match so "--pr" is not satisfied by "--propagate-tags".
	assertExactFlagCandidates(t, resp.Stdout, "--pr", "--title", "--comment")
	// Removed flags must not appear as completion candidates.
	for _, bad := range []string{"--dep", "--all-deps"} {
		for _, line := range strings.Split(resp.Stdout, "\n") {
			if strings.TrimSpace(line) == bad {
				t.Fatalf("completion must not offer %s; stdout=%q", bad, resp.Stdout)
			}
		}
	}
	assertNoEventsJSONL(t, resp)
}
```
