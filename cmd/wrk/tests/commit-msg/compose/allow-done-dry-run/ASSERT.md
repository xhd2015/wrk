## Expected

- Flag layer accepts `--commit -m … --done --dry-run` (flags recognized).
- Stderr must **not** contain `mutually exclusive`.
- Stderr must **not** contain `--dry-run is only valid with` (dry-run host rejection).
- Stderr must **not** be `unrecognized flag` for `--commit` / `-m` / `--message`.

## Side Effects

- None required for flag-matrix leaf.

## Exit Code

- Any, as long as failure is not mutex, invalid dry-run host, or unrecognized flags.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "unrecognized flag") {
		t.Fatalf("flag layer does not recognize --commit/-m yet (unrecognized flag); stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("--commit -m --done --dry-run rejected as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "--dry-run is only valid with") {
		t.Fatalf("compose dry-run still rejected as invalid dry-run host; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
