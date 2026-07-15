## Expected Output

```
merged branch <WtBranch> into main

feature-stays ← main  (+1 commit)

synced: 0 into main, 1 into worktrees, 0 skipped

v0.0.1        owned changed                  ->  v0.0.2
tagged v0.0.2 @ <short>
1 tag created

pushed main → origin/main
```

## Expected

- Exit code 0.
- Fixed order: primary → sync → tag-next → push (blank line between major stages).
- Stderr empty.
- wtA removed; wtB HEAD == main; `feature-work` on main.
- Local + origin `v0.0.2`; origin/main == main HEAD.
- Event command `"done"`.

## Side Effects

- Remaining worktree synced; tags created; branch+tags pushed.

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

	primary := strings.TrimSuffix(primaryMergeMsg(req.WtBranch), "\n")
	syncBlock := buildSyncStdout([]string{syncDetailPass2(req.Wt2Branch, 1)}, 0, 1, 0)
	short := shortHEAD(t, req.MainRepo)
	want := joinMajorStages(
		primary,
		syncBlock,
		tagNextRootBumpApplyStdout(short),
		donePushConfirmLine(),
	)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertHEAD(t, req.Wt2Dir, revParseHEAD(t, req.MainRepo))

	assertLocalTagAtMainHEAD(t, req.MainRepo, "v0.0.2")
	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, "v0.0.2") {
		t.Fatal("v0.0.2 should exist on bare origin after full pipeline")
	}
	assertLastEventCommandDone(t, req.WrkHome)
}
```
