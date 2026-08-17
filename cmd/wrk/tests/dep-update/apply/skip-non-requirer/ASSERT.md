## Expected

- Exit 0.
- Pin + tidy for `example.com/app` only.
- No pin/tidy line for `example.com/other`.
- Sibling go.mod identical to baseline (no new require).

## Side Effects

- Only modules that already required xxx are mutated.

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
	assertNotContains(t, resp.Stdout, req.WantConsumer2Module)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertGoSumExists(t, req.ConsumerModDir)
}
```
