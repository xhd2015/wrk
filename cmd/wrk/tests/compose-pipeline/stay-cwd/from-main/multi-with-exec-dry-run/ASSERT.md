
## Expected

- Flag layer accepts multi-stage + `--exec` without `--done`.
- Stderr must not contain `mutually exclusive`.
- Stderr must not restrict `--exec` to only `--done`/create/cd modes for this compose path.
- Exit 0 preferred with tag plan present.

## Side Effects

- None required (dry-run).

## Exit Code

- 0 preferred; failure must not be flag rejection of this combo.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	assertNoMutexReject(t, se)
	if strings.Contains(se, "--exec is not valid") ||
		strings.Contains(se, "--exec is only valid") ||
		(strings.Contains(se, "--exec") && strings.Contains(se, "not valid with")) {
		t.Fatalf("--exec rejected with multi-stage without done; stderr=%q", se)
	}
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("multi-stage+exec still exclusive; stderr=%q", se)
	}
	if resp.ExitCode != 0 && (strings.Contains(se, "mutually exclusive") ||
		strings.Contains(se, "only valid") ||
		strings.Contains(se, "not valid")) {
		t.Fatalf("exit %d flag-policy reject for multi+exec; stderr=%q", resp.ExitCode, se)
	}
	if resp.ExitCode == 0 {
		if !strings.Contains(resp.Stdout, "1 tag planned") {
			t.Fatalf("expected tag-next plan on success path; stdout=%q", resp.Stdout)
		}
	}
}
```
