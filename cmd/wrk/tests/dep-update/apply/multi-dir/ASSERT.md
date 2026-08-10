## Expected

- Exit 0.
- Stdout dep-update lines for both modules with WantVersion / WantVersion2.
- Both replaces dropped; requires at latest tags.
- No tidy.

## Side Effects

- Two modules unpinned from local replace to tagged requires.

## Exit Code

- 0

```go
import (
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
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.ConsumerGoMod, modDep2, req.WantVersion2)
	assertNoTidyArtifacts(t, req)
}
```
