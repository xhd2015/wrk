
## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

updated example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.2.3
  go build ./... ok
  committed <short7>  chore(deps): bump example.com/lib to v1.2.3

updated 1 module across 1 project
```

## Expected

- Exit code 0.
- Stdout uses apply verbs (not `would:`); includes `go build ./... ok` and
  `committed <short7>  chore(deps): bump example.com/lib to v1.2.3` where short7
  matches `git rev-parse --short=7 HEAD` on the consumer after the run.
- Footer `updated 1 module across 1 project`.
- Stderr empty.

## Side Effects

- App HEAD advanced by one commit; parent is pre-run HEAD.
- Commit subject exactly `chore(deps): bump example.com/lib to v1.2.3`.
- Commit paths are only `go.mod` and/or `go.sum`.
- App `go.mod` require for `example.com/lib` is `v1.2.3`.
- Source go.mod / tags / HEAD unchanged; app tags unchanged.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "would:") {
		t.Fatalf("apply stdout must not use would: prefix, got %q", resp.Stdout)
	}

	subject := depsBumpSubject("example.com/lib", "v1.2.3")
	assertAppDepsCommitted(t, req, subject)
	assertApplySourceUnchanged(t, req)

	short := shortHEAD(t, req.AppPath)
	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib", "v1.2.3", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(updatedHeader("example.com/app", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine("example.com/lib", "v1.0.0", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteString(goBuildOkLine())
	b.WriteByte('\n')
	b.WriteString(committedLine(short, subject))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(applyFooter(1, 1))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	gotMod := readFile(t, goModPath(req.AppPath))
	assertGoModRequireVersion(t, gotMod, "example.com/lib", "v1.2.3")
}
```
