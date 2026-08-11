## Expected Output

```
dep-update example.com/lib -> v1.2.3
go mod tidy ok  module example.com/app
dep-update: updated 1, already 0, skipped 0
```

## Expected

- Exit 0.
- Apply pin line for `example.com/lib` → `v1.2.3` (optional tag parenthetical OK).
- `go mod tidy ok` for `example.com/app`.
- Summary `dep-update: updated 1, already 0, skipped 0`.
- No `would:` vocabulary.
- Trailing newline.

## Side Effects

- Consumer require bumped to v1.2.3; go.sum exists after tidy.
- Owner go.mod unchanged (blast radius: consumer toplevel only).
- No git commit required.

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
	assertDepUpdateLine(t, resp.Stdout, modLib, req.WantVersion)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, false)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertOwnerGoModUnchanged(t, req)
}
```
