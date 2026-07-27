
## Expected Output

Three major stdout stages (tag-next → push confirm → propagate), blank-separated:

```
v1.0.0        owned changed                  ->  v1.0.1
tagged v1.0.1 @ <source-short>
1 tag created

pushed main → origin/main

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
- Local tag `v1.0.1` at source HEAD.
- Bare origin has `refs/tags/v1.0.1` **and** `refs/heads/main` == source HEAD.
- Stdout includes `pushed main → origin/main` between tag-next and propagate stages.
- Consumer require bumped and committed (same as tag-then-propagate).

## Side Effects

- Source HEAD unchanged; new local + remote tag; branch tip on origin.
- App go.mod at `v1.0.1`; one deps commit.

## Exit Code

- 0

```go
import (
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
		t.Fatalf("apply compose must not use would: prefix, got %q", resp.Stdout)
	}

	if req.OriginBare == "" {
		t.Fatal("OriginBare empty")
	}
	if !tagRefExists(t, req.SourcePath, req.NextTag) {
		t.Fatalf("%s should exist locally after --push compose", req.NextTag)
	}
	if !remoteTagExists(t, req.OriginBare, req.NextTag) {
		t.Fatalf("%s should exist on bare origin after --push compose", req.NextTag)
	}
	// Branch tip published with tags (not tags-only).
	srcHEAD := headSHA(t, req.SourcePath)
	originMain := strings.TrimSpace(gitOutputIsolated(t, req.OriginBare, "rev-parse", "refs/heads/main"))
	if originMain != srcHEAD {
		t.Fatalf("origin/main %s != source HEAD %s (branch must be pushed)", originMain, srcHEAD)
	}
	assertSourceHEADUnchanged(t, req)

	subject := depsBumpSubject(req.ModulePath, req.NextTag)
	assertAppDepsCommitted(t, req, subject)
	assertGoModRequireVersion(t, readFile(t, goModPath(req.AppPath)), req.ModulePath, req.NextTag)

	srcShort := shortHEAD(t, req.SourcePath)
	appShort := shortHEAD(t, req.AppPath)
	pushLine := "pushed main → origin/main\n"
	want := joinMajorStages(
		tagNextRootBumpApplyStdout(req.OldTag, req.NextTag, srcShort),
		pushLine,
		propStageApplyStdout(req, appShort, subject),
	)
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
}
```
