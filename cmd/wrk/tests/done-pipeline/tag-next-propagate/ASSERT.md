## Expected Output

Three major stages separated by blank lines (primary → tag-next → propagate):

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

- Exit code 0.
- Stderr empty.
- Fixed order: primary → tag-next → propagate-tags (blank line between stages).
- Worktree removed; local tag `v0.0.2` at main HEAD.
- App require bumped to `v0.0.2` and committed with `chore(deps):` subject.
- Last `events.jsonl` event has `command: "done"` (not `"propagate-tags"` / `"tag-next"`).

## Side Effects

- Source gets new lightweight tag; consumer go.mod/go.sum committed; source wt gone.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	// Must not fail as mutual exclusion (P7 flag-layer unlock).
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("flag layer still rejects done+propagate-tags; stderr=%q", resp.Stderr)
	}

	srcShort := shortHEAD(t, req.MainRepo)
	appShort := shortHEAD(t, req.SecondRepo)
	subject := propDepsBumpSubject(req.DepModulePath, pipelinePropNextTag)
	want := joinMajorStages(
		primaryMergeMsg(req.WtBranch),
		tagNextRootBumpApplyStdout(srcShort),
		propStageApplyStdout(req.MainRepo, req.DepModulePath, pipelinePropOldTag, pipelinePropNextTag,
			filepath.Base(req.SecondRepo), appShort, subject),
	)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertLocalTagAtMainHEAD(t, req.MainRepo, pipelinePropNextTag)
	assertAppBumpedAndCommitted(t, req, pipelinePropNextTag, subject)
	assertLastEventCommandDone(t, req.WrkHome)
}
```
