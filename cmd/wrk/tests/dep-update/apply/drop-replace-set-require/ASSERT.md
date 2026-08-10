## Expected

- Exit 0.
- Stdout `dep-update example.com/dep -> v0.0.2` (optional tag form OK).
- No `would:` vocabulary.
- go.mod: no replace for example.com/dep; require at v0.0.2.
- No go.sum (D2 no tidy).

## Side Effects

- Replace dropped; require pinned to latest tag version.

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
	assertNotContains(t, resp.Stdout, "would:")
	assertDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertNoTidyArtifacts(t, req)
}
```
