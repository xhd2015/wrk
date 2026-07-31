## Expected

- Non-zero exit.
- Stderr mentions **no open pull request** (case-insensitive substring match on those words).
- Preferably stderr also names the head branch (`feature-pr`).
- No PR success tokens on stdout (`comment added` / `PR created`).
- Fake `gh`: **`pr create` and `pr comment` not called**; `pr list` may run.
- No ensure-push / `pushed` line.

## Errors

- Comment-only never creates a PR. Missing open PR is a hard error (unlike bare show, which prints empty and exits 0).

## Side Effects

- No git push; no PR create; no issue comment.

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
		t.Fatalf("expected non-zero when comment-only and no open PR; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}

	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "no open") || !strings.Contains(se, "pull request") {
		t.Fatalf("stderr should mention no open pull request, got %q", resp.Stderr)
	}
	// Prefer naming the head branch (soft — contract says preferably).
	_ = strings.Contains(resp.Stderr, req.WtBranch) || strings.Contains(resp.Stderr, prFeatureBranch)

	for _, tok := range []string{"comment added", "PR created", "title set", "body set", "pushed "} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
	_ = invocs
}
```
