## Expected

- Non-zero exit.
- Stderr mentions **no open pull request** (case-insensitive substring match on those words).
- Preferably stderr also names the head branch (`feature-pr`).
- No PR success tokens on stdout (`pushed `, `comment added`, `PR created`).
- Fake `gh`: **`pr create` and `pr comment` not called**; `pr list` may run.
- Origin `refs/heads/feature-pr` still equals the **pre-run snapshot** (no push).

## Errors

- Push-existing never creates a PR. Missing open PR is a hard error (unlike bare show, which prints empty and exits 0). List must run **before** push so origin tip stays unchanged.

## Side Effects

- No git push; no PR create; no issue comment.

## Exit Code

- Non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when push-existing and no open PR; stdout=%q stderr=%q",
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

	// Load-bearing: list-before-push — origin tip must not advance.
	beforeBytes, readErr := os.ReadFile(filepath.Join(req.WorkRoot, "origin-feature-before"))
	if readErr != nil {
		t.Fatalf("read origin snapshot: %v", readErr)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/"+req.WtBranch)
	if after != before {
		t.Fatalf("origin/%s must stay unchanged when no open PR: before %s after %s",
			req.WtBranch, before, after)
	}
	// Fixture still has local ahead of origin.
	local := revParseHEAD(t, req.WtDir)
	if local == after {
		t.Fatal("fixture expected local HEAD ahead of origin tip")
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
	_ = invocs
}
```
