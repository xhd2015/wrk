## Expected Output

Primary merge message, blank line, then tag-next apply block (short hash runtime):

```
merged branch <WtBranch> into main

v0.0.1        owned changed                  ->  v0.0.2
tagged v0.0.2 @ <short>
1 tag created
```

## Expected

- Exit code 0.
- Stdout: primary merge message, blank line, root-bump tag-next apply block.
- Stderr empty.
- Worktree directory gone; branch deleted; main has `feature-work`.
- Lightweight tag `v0.0.2` exists locally at main HEAD (no remote required).
- Last `events.jsonl` event has `command: "done"` (not `"tag-next"`).

## Side Effects

- Local tag created; no push side effects required.

## Exit Code

- 0

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	short := shortHEAD(t, req.MainRepo)
	want := joinMajorStages(
		primaryMergeMsg(req.WtBranch),
		tagNextRootBumpApplyStdout(short),
	)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))

	assertLocalTagAtMainHEAD(t, req.MainRepo, "v0.0.2")
	assertLastEventCommandDone(t, req.WrkHome)
}
```
