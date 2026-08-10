## Expected

- Exit **0** (soft tidy failures do not fail the process).
- Stderr contains `warning:` mentioning go mod tidy (and preferably the bad module dir).
- Root consumer go.mod has relative replace for dep (continue-after-fail proven).
- Bad module may also have replace written before tidy failed.
- Summary: applied >= 1, tidy failed >= 1.

## Side Effects

- At least the good consumer is pinned despite peer tidy failure.

## Exit Code

- 0

```go
import (
	"fmt"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertWarningTidy(t, resp.Stderr)
	assertRelativeReplace(t, req.ConsumerGoMod, modDep)
	out := resp.Stdout
	if !strings.Contains(out, "pin-local "+modConsumer) && !strings.Contains(out, "pin-local "+modBad) {
		t.Fatalf("expected at least one pin-local line for consumer or bad; stdout:\n%s", out)
	}
	if !strings.Contains(out, "tidy failed") {
		t.Fatalf("summary must mention tidy failed; stdout:\n%s", out)
	}
	ok := false
	for applied := 1; applied <= 2; applied++ {
		for f := 1; f <= 2; f++ {
			for tok := 0; tok <= 2; tok++ {
				want := fmt.Sprintf("pin-locals: applied %d, tidy ok %d, tidy failed %d", applied, tok, f)
				if strings.Contains(out, want) {
					ok = true
				}
			}
		}
	}
	if !ok {
		assertContains(t, out, "applied")
		assertContains(t, out, "tidy failed")
		t.Fatalf("want locked summary pin-locals: applied N, tidy ok M, tidy failed F (F>=1); stdout:\n%s", out)
	}
}
```
