## Expected

- Exit 0.
- Stdout dep-update lines for both modules with WantVersion / WantVersion2.
- One `go mod tidy ok  module example.com/consumer`.
- Both replaces dropped; requires at latest tags; go.sum exists.

## Side Effects

- Two modules unpinned from local replace to tagged requires.

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
	assertExitZero(t, resp)
	assertDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	assertDepUpdateLine(t, resp.Stdout, modDep2, req.WantVersion2)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep2)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.ConsumerGoMod, modDep2, req.WantVersion2)
	assertGoSumExists(t, req.ConsumerModDir)
	if strings.Count(resp.Stdout, "go mod tidy ok") != 1 {
		t.Fatalf("expected tidy once for one consumer; got:\n%s", resp.Stdout)
	}
}
```
