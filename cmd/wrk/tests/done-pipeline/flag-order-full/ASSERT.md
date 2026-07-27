## Expected

- Exit code 0.
- Same ordered composition stdout as `sync-tag-next-push` (primary → sync → tag-next → push).
- Same side effects: wtA gone; wtB synced; local+origin tags; origin/main == main.
- Event command `"done"`.

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
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, "v0.0.2") {
		t.Fatal("v0.0.2 should exist on bare origin")
	}
	assertLastEventCommandDone(t, req.WrkHome)
}
```
