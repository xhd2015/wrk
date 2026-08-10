## Expected

- Exit 0.
- No replace for `example.com/other` in consumer go.mod.
- No pin-local line naming other as dep.
- Dep may be pinned (required).

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
	assertNoReplaceFor(t, req.ConsumerGoMod, modOther)
	assertNotContains(t, resp.Stdout, "<- "+modOther)
	// Dep is required — should be pinned when product exists.
	if strings.Contains(resp.Stdout, "pin-local ") {
		assertPinLocalLine(t, resp.Stdout, modConsumer, modDep)
		assertRelativeReplace(t, req.ConsumerGoMod, modDep)
	}
}
```
