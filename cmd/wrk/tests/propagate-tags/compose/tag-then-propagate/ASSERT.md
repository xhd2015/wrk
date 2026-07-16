## Expected Output

Two major stages separated by a blank line:

```
v1.0.0        owned changed                  ->  v1.0.1
tagged v1.0.1 @ <source-short>
1 tag created

source: /abs/.../lib
  example.com/lib  @ v1.0.1  (tag v1.0.1)

updated example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.0.1
  go build ./... ok
  committed <app-short7>  chore(deps): bump example.com/lib to v1.0.1

updated 1 module across 1 project
```

## Expected

- Exit code 0.
- Stderr empty.
- Multi-stage stdout: tag-next apply block, blank line, propagate apply block.

## Side Effects

- Lightweight tag `v1.0.1` exists on source at HEAD.
- Source HEAD and go.mod unchanged (only new tag ref).
- App `go.mod` require for `example.com/lib` is `v1.0.1`.
- App HEAD advances one commit with subject
  `chore(deps): bump example.com/lib to v1.0.1` (go.mod/go.sum only).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "would:") {
		t.Fatalf("apply compose must not use would: prefix, got %q", resp.Stdout)
	}

	if !tagRefExists(t, req.SourcePath, req.NextTag) {
		t.Fatalf("%s tag should exist after compose apply", req.NextTag)
	}
	gotTag := gitOutputIsolated(t, req.SourcePath, "rev-parse", req.NextTag)
	head := gitOutputIsolated(t, req.SourcePath, "rev-parse", "HEAD")
	if gotTag != head {
		t.Fatalf("%s should point at HEAD: tag=%s head=%s", req.NextTag, gotTag, head)
	}
	assertSourceHEADUnchanged(t, req)

	subject := depsBumpSubject(req.ModulePath, req.NextTag)
	assertAppDepsCommitted(t, req, subject)

	gotMod := readFile(t, goModPath(req.AppPath))
	assertGoModRequireVersion(t, gotMod, req.ModulePath, req.NextTag)

	srcShort := shortHEAD(t, req.SourcePath)
	appShort := shortHEAD(t, req.AppPath)
	want := joinMajorStages(
		tagNextRootBumpApplyStdout(req.OldTag, req.NextTag, srcShort),
		propStageApplyStdout(req, appShort, subject),
	)
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
}
```
