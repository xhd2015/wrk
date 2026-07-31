## Expected

- Non-zero exit.
- Stderr indicates invalid combination of PR status with create flags (mutually exclusive / not valid / only valid / cannot / invalid).
- Prefer mentioning `--status` and/or `--title` when product wording allows.
- No status stdout block (`State:` / `Checks:`).
- Fake `gh`: **`pr create` not called** (fail before create path).

## Errors

- `--pr --status` is read-only; cannot compose with `--title` / `--comment` create/attach.

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
		t.Fatalf("expected non-zero for --pr --status --title --comment; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "only valid") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "invalid") {
		t.Fatalf("stderr should indicate invalid combination / exclusion, got %q", resp.Stderr)
	}
	for _, tok := range []string{"State:", "Checks:", "PR created", "comment added"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("refuse must not print status/create tokens %q; stdout=%q", tok, resp.Stdout)
		}
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	_ = invocs
}
```
