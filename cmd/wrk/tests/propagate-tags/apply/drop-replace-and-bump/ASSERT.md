## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

updated example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.2.3
  go build ./... ok
  committed <short7>  chore(deps): bump example.com/lib to v1.2.3

dropped replace example.com/lib  (project app)

updated 1 module across 1 project
```

## Expected

- Exit code 0.
- Stdout includes version bump, `go build ./... ok`, `committed … chore(deps):…`,
  and `dropped replace` (not `would:`).
- Footer `updated 1 module across 1 project`.

## Side Effects

- App `go.mod` require for `example.com/lib` is `v1.2.3`.
- App `go.mod` has **no** `replace` for `example.com/lib`.
- **P5:** app HEAD advances with deps commit (go.mod/go.sum only).
- Source go.mod / tags / HEAD unchanged; app tags unchanged.

## Justification (contract expansion)

Same as root-bump: P5 adds build gate + commit on successful apply. Drop-replace
still reports after the update block; build/commit lines sit under the update
block after version arrows.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "would:") {
		t.Fatalf("apply stdout must not use would: prefix, got %q", resp.Stdout)
	}

	// Fixture must have started with a replace.
	if !strings.Contains(req.AppGoModBefore, "replace example.com/lib") {
		t.Fatalf("fixture go.mod missing replace; got %q", req.AppGoModBefore)
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
	b.WriteString(droppedReplaceLine("example.com/lib", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(applyFooter(1, 1))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	gotMod := readFile(t, goModPath(req.AppPath))
	assertGoModRequireVersion(t, gotMod, "example.com/lib", "v1.2.3")
	assertGoModNoReplace(t, gotMod, "example.com/lib")
}
```
