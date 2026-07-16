## Expected Output

stdout (plain):

```
would: go run ./script/foo/install
would: reinstall 1 binaries (0 skipped)
```

stderr (prefix colored grey `#90`):

```
notice: bin foo: preferring ./script/foo/install over ./cmd/foo
```

## Expected

- Exit code 0.
- Stdout exact plan lines with **no** ANSI.
- Stderr: only the `notice:` token wrapped in grey ANSI; remainder plain.
- Stub binary unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	wantOut := "would: go run ./script/foo/install\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(wantOut))
	assertNoANSI(t, resp.Stdout)
	assertNotContains(t, resp.Stdout, "go install ./cmd/foo")
	wantErr := coloredNoticePrefix() + " bin foo: preferring ./script/foo/install over ./cmd/foo\n"
	assert.Output(t, resp.Stderr, v2StdoutTemplate(wantErr))
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
