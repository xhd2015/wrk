## Expected

- Non-zero exit.
- Stderr mentions **no open pull request** (case-insensitive on those words).
- Preferably stderr also names the head branch (`feature-pr`).
- No success tokens on stdout (`pushed `, `comment added`, `PR created`).
- Fake `gh`: **`pr create` and `pr comment` not called**; `pr list` may run.
- Origin tip still equals pre-run snapshot (no push).

## Errors

- Same open-PR gate as bare push-existing; comment is never applied when no open PR.

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
		t.Fatalf("expected non-zero when push+comment and no open PR; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}

	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "no open") || !strings.Contains(se, "pull request") {
		t.Fatalf("stderr should mention no open pull request, got %q", resp.Stderr)
	}
	_ = strings.Contains(resp.Stderr, req.WtBranch) || strings.Contains(resp.Stderr, prFeatureBranch)

	for _, tok := range []string{"comment added", "PR created", "title set", "body set", "pushed "} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

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
