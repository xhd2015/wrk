
## Expected Output

stdout (plain):

```
would: reinstall 0 binaries (0 skipped)
```

stderr (prefix colored orange `#33`):

```
warning: bin foo: ambiguous under cmd (./cmd/foo, ./cmd/nested/foo); skipping
```

## Expected

- Exit code 0.
- Stdout exact summary with **no** ANSI.
- Stderr: only the `warning:` token wrapped in orange ANSI; remainder plain.
- Stub binary unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	wantOut := "would: reinstall 0 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(wantOut))
	assertNoANSI(t, resp.Stdout)
	wantErr := coloredWarningPrefix() +
		" bin foo: ambiguous under cmd (./cmd/foo, ./cmd/nested/foo); skipping\n"
	assert.Output(t, resp.Stderr, v2StdoutTemplate(wantErr))
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
