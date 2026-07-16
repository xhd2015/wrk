## Expected Output

```
merged branch <WtBranch> into main

v0.0.1        owned changed                  ->  v0.0.2
tagged v0.0.2 @ <short>
1 tag created

source: <MainRepo>
  example.com/lib  @ v0.0.2  (tag v0.0.2)

updated example.com/app  (project app)
  example.com/lib  v0.0.1 -> v0.0.2
  go build ./... ok
  committed <app-short7>  chore(deps): bump example.com/lib to v0.0.2

updated 1 module across 1 project
```

## Expected

- Exit code 0; stderr empty.
- Source worktree **kept** (no `worktree removed:`).
- Tag `v0.0.2` at main HEAD; app bumped and committed.
- Event command `"merge-back"` (not `"done"` / `"propagate-tags"`).

## Side Effects

- Local tag + consumer commit; source wt/branch remain.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("flag layer still rejects merge-back+propagate-tags; stderr=%q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "worktree removed:")

	srcShort := shortHEAD(t, req.MainRepo)
	appShort := shortHEAD(t, req.SecondRepo)
	subject := mbPropDepsBumpSubject(req.DepModulePath, mbPipelinePropNextTag)
	want := joinMajorStages(
		primaryMergeMsg(req.WtBranch),
		tagNextRootBumpApplyStdout(srcShort),
		mbPropStageApplyStdout(req.MainRepo, req.DepModulePath, mbPipelinePropOldTag, mbPipelinePropNextTag,
			filepath.Base(req.SecondRepo), appShort, subject),
	)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertSourceWorktreeKept(t, req)
	assertLocalTagAtMainHEAD(t, req.MainRepo, mbPipelinePropNextTag)
	mbAssertAppBumpedAndCommitted(t, req, mbPipelinePropNextTag, subject)
	assertLastEventCommandMergeBack(t, req.WrkHome)
}
```
