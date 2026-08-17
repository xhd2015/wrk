## Expected

- Exit 0.
- Pin + `go mod tidy ok` for `example.com/consumer`.
- go.sum exists; require @ latest.
- Wrapper at `go1.19.13` ran (recorded GOROOT/PATH0 contains that pin).

## Side Effects

- Tidy invoked via withgo pin `go1.19.13`, not the host default SDK.

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
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertVersionedGoUsed(t, req)
}
```
