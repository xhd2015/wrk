## Expected

- Exit **0**.
- Stderr: `warning: skipping nested checkout sandbox/broken-wt` (path-qualified).
- Stderr reason mentions not a git repository / nonexistent gitdir.
- No leaked raw `fatal:` as the sole failure (command must not hard-abort).
- Dry-run banner + `would: pin` for `example.com/dep`; summary would update 1.
- go.mod unchanged.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertWarningStderr(t, resp.Stderr)
	nested := req.WantBrokenNested
	if nested == "" {
		nested = "sandbox/broken-wt"
	}
	se := resp.Stderr
	if !strings.Contains(se, "skipping nested checkout") {
		t.Fatalf("stderr must say skipping nested checkout, got %q", se)
	}
	// Prefer relative display; allow slash-normalized match.
	wantRel := filepath.ToSlash(nested)
	if !strings.Contains(filepath.ToSlash(se), wantRel) {
		t.Fatalf("stderr must include nested path %q, got %q", wantRel, se)
	}
	low := strings.ToLower(se)
	if !strings.Contains(low, "not a git repository") && !strings.Contains(low, "gitdir") {
		t.Fatalf("warning reason should mention broken git repo, got %q", se)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update \(dry-run\) ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-update: would update 1 modules in 1 checkouts
`)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
