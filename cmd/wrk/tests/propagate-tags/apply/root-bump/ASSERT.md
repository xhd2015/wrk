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
- Stdout matches apply shape: `updated` (not `would:`) plus P5 additive
  `go build ./... ok` and `committed <short7>  chore(deps): …`.
- Source block lists the root release.
- Footer `updated 1 module across 1 project`.
- Stderr empty.

## Side Effects

- App `go.mod` require for `example.com/lib` is `v1.2.3` (bumped from `v1.0.0`).
- Source `go.mod` unchanged.
- **P5:** app HEAD advances with `chore(deps): bump example.com/lib to v1.2.3`;
  commit paths only go.mod/go.sum. Source HEAD/tags and app tags unchanged.

## Justification (contract expansion)

P5 intentionally adds build gate + commit after successful tidy. This leaf previously
asserted no commit (P4). Updating expects the additive success lines and HEAD move.

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
	if strings.Contains(gotMod, "v1.0.0") {
		for _, line := range strings.Split(gotMod, "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) >= 2 && fields[0] == "example.com/lib" && fields[1] == "v1.0.0" {
				t.Fatalf("app go.mod still requires v1.0.0:\n%s", gotMod)
			}
			if len(fields) >= 3 && fields[0] == "require" && fields[1] == "example.com/lib" && fields[2] == "v1.0.0" {
				t.Fatalf("app go.mod still requires v1.0.0:\n%s", gotMod)
			}
		}
	}
}
```
