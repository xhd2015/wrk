## Expected

- Exit 0.
- Stdout contains `dep-replace example.com/dep => <abs>`.
- go.mod has absolute replace for `example.com/dep` (not `./` / `../`).
- No `would:` vocabulary.
- No go.sum created (D2 no tidy).

## Side Effects

- Consumer go.mod gains absolute replace to dep dir.

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
	assertDepReplaceLine(t, resp.Stdout, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertNoTidyArtifacts(t, req)
}
```
