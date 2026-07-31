## Expected

- Exit code **0** (soft no-open; not an error path like comment-only).
- Stdout **empty** (no bytes).
- Stderr contains `warning:` and mentions no open pull request (`no open` + `pull request`, case-insensitive).
- Preferably stderr also names the head branch (`feature-pr`).
- Fake `gh`: **`pr view`, `pr create`, and `pr comment` not called**; `pr list` may run.
- No create/attach/push tokens on stdout.

## Side Effects

- No git push; no PR create; no issue comment; no `pr view`.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("no open PR for --pr --status must exit 0; got %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != "" {
		t.Fatalf("no open PR: stdout must be empty; got %q", resp.Stdout)
	}

	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "warning:") {
		t.Fatalf("stderr should contain warning: prefix, got %q", resp.Stderr)
	}
	if !strings.Contains(se, "no open") || !strings.Contains(se, "pull request") {
		t.Fatalf("stderr should mention no open pull request, got %q", resp.Stderr)
	}
	// Prefer naming the head branch (soft).
	_ = strings.Contains(resp.Stderr, req.WtBranch) || strings.Contains(resp.Stderr, prFeatureBranch)

	for _, tok := range []string{"PR created", "comment added", "title set", "body set", "pushed ", "State:", "Checks:"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "view")
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
	// list is allowed (and expected after implementer wires status).
	_ = invocs
}
```
