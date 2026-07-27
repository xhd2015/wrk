
## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

would: update example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.2.3

would: drop replace example.com/lib  (project app)

would: update 1 module across 1 project
```

## Expected

- Exit code 0.
- Plan includes version bump **and** `would: drop replace example.com/lib  (project app)`.
- Footer `would: update 1 module across 1 project`.

## Side Effects

- App `go.mod` still contains the old require version and the replace directive (byte-identical to pre-run).
- Source/app tags and HEAD unchanged.

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

	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib", "v1.2.3", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(wouldUpdateHeader("example.com/app", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine("example.com/lib", "v1.0.0", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(dropReplaceLine("example.com/lib", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(planFooter(1, 1))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	assertDryRunNoMutation(t, req)
	// Explicit: replace line still present in go.mod after dry-run.
	if !strings.Contains(req.AppGoModBefore, "replace example.com/lib") {
		t.Fatalf("fixture go.mod missing replace; got %q", req.AppGoModBefore)
	}
	gotMod := readFile(t, goModPath(req.AppPath))
	if !strings.Contains(gotMod, "replace example.com/lib") {
		t.Fatalf("dry-run must not drop replace from go.mod; after: %q", gotMod)
	}
	if !strings.Contains(gotMod, "v1.0.0") {
		t.Fatalf("dry-run must not bump require version; after: %q", gotMod)
	}
}
```
