## Expected Output

Primary merge message, blank line, then propagate apply (no tag-next block):

```
merged branch <WtBranch> into main

source: <MainRepo>
  example.com/lib  @ v0.0.2  (tag v0.0.2)

updated example.com/app  (project app)
  example.com/lib  v0.0.1 -> v0.0.2
  go build ./... ok
  committed <app-short7>  chore(deps): bump example.com/lib to v0.0.2

updated 1 module across 1 project
```

## Expected

- Exit code 0.
- Stderr empty.
- Stdout has **no** `tag created` / `1 tag planned` / `tagged ` (tag-next not requested).
- App bumped to existing `v0.0.2` and committed.
- Worktree removed; event command `"done"`.

## Side Effects

- Consumer commit only; no new source tags beyond pre-seeded `v0.0.2`.

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
		t.Fatalf("flag layer still rejects done+propagate-tags; stderr=%q", resp.Stderr)
	}

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "tag planned")
	assertNotContains(t, resp.Stdout, "tagged ")

	appShort := shortHEAD(t, req.SecondRepo)
	subject := propDepsBumpSubject(req.DepModulePath, pipelinePropNextTag)
	want := joinMajorStages(
		primaryMergeMsg(req.WtBranch),
		propStageApplyStdout(req.MainRepo, req.DepModulePath, pipelinePropOldTag, pipelinePropNextTag,
			filepath.Base(req.SecondRepo), appShort, subject),
	)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertAppBumpedAndCommitted(t, req, pipelinePropNextTag, subject)
	assertLastEventCommandDone(t, req.WrkHome)
}
```
