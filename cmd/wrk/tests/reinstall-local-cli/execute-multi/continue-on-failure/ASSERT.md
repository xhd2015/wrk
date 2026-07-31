
## Expected Output

```
go install ./cmd/broken
go install ./cmd/toolgood
reinstalled 1, skipped 0, failed 1
```

(Child `go` compiler diagnostics may appear on stderr between progress lines;
stdout wrk-owned lines are the three above when implementer does not mirror
compiler noise onto stdout.)

## Expected

- Exit code **0** (soft: install failures never fail the process; `failed > 0`
  yields warning + exit 0, not ExitCodeError 1).
- Stderr contains a `warning:` line that mentions reinstall and/or fail
  (exact wording implementer-owned; child `go` noise may also appear).
- Stdout contains progress for both packages in multi-plan order (`broken` in
  root module before `toolgood` in nested `tools/`).
- Last content line of stdout is exactly `reinstalled 1, skipped 0, failed 1`.
- No `would:` dry-run vocabulary and no `# module` / `across` dry-run headers.
- `$GOBIN/toolgood` is installed (not stub) and prints `toolgood-ok` — prove
  continue-on-failure **across modules** (later module still ran after earlier
  module install failure).
- Failed count in summary is exactly 1.

## Side Effects

- `go install ./cmd/broken` fails (compile error); stub for broken may remain or
  be partially written — not asserted.
- `go install ./cmd/toolgood` still runs after the failure (different ModuleRoot)
  and replaces the toolgood stub.

## Errors

- Overall process finishes the multi plan (not abort-on-first-error / not skip
  remaining modules) and exits **0** with a soft `warning:` when `failed > 0`.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if !strings.Contains(resp.Stderr, "warning:") {
		t.Fatalf("stderr must contain warning: when failed > 0, got %q", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "reinstall") && !strings.Contains(lower, "fail") {
		t.Fatalf("stderr warning must mention reinstall or fail, got %q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "would:")
	assertNotContains(t, resp.Stdout, "# module")
	assertNotContains(t, resp.Stdout, "across")
	assertContains(t, resp.Stdout, "go install ./cmd/broken")
	assertContains(t, resp.Stdout, "go install ./cmd/toolgood")
	// root module (broken) before nested tools (toolgood)
	iBroken := strings.Index(resp.Stdout, "go install ./cmd/broken")
	iGood := strings.Index(resp.Stdout, "go install ./cmd/toolgood")
	if iBroken < 0 || iGood < 0 || iBroken > iGood {
		t.Fatalf("want broken progress before toolgood; stdout=%q", resp.Stdout)
	}
	assertExecuteSummary(t, resp.Stdout, 1, 0, 1)
	assertBinNotStub(t, req.BinDir, "toolgood")
	assertBinExecutable(t, req.BinDir, "toolgood")
	assertBinRuns(t, req.BinDir, "toolgood", "toolgood-ok")
}
```
