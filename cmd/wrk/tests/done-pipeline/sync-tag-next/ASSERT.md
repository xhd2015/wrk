## Expected Output

```
merged branch <WtBranch> into main

feature-stays ← main  (+1 commit)

synced: 0 into main, 1 into worktrees, 0 skipped

v0.0.1        owned changed                  ->  v0.0.2
tagged v0.0.2 @ <short>
1 tag created
```

## Expected

- Exit code 0.
- Ordered stdout: primary → sync block → tag-next apply (blank lines between major stages).
- No push confirmation line.
- wtA removed; wtB HEAD == main HEAD; local `v0.0.2` at main HEAD.
- Event command `"done"`.

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
	want := joinMajorStages(primary, syncBlock, tagNextRootBumpApplyStdout(short))
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	if strings.Contains(resp.Stdout, "pushed main → origin/main") {
		t.Fatalf("stdout must not include push line without --push; got %q", resp.Stdout)
	}

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertHEAD(t, req.Wt2Dir, revParseHEAD(t, req.MainRepo))

	assertLocalTagAtMainHEAD(t, req.MainRepo, "v0.0.2")
	assertLastEventCommandDone(t, req.WrkHome)
}
```
