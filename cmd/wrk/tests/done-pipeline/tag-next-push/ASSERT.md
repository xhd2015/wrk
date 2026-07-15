## Expected Output

```
merged branch <WtBranch> into main

v0.0.1        owned changed                  ->  v0.0.2
tagged v0.0.2 @ <short>
1 tag created

pushed main → origin/main
```

## Expected

- Exit code 0.
- Ordered stdout: primary → blank → tag-next apply → blank → push confirmation.
- Stderr empty.
- Worktree removed; main has `feature-work`.
- Local `v0.0.2` at main HEAD.
- Bare origin `refs/heads/main` == main HEAD; `refs/tags/v0.0.2` exists on origin.
- Last event `command: "done"`.

## Side Effects

- Branch tip and new tag published to origin via `runPushMain(..., createdTags)`.

## Exit Code

- 0

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	short := shortHEAD(t, req.MainRepo)
	want := joinMajorStages(
		primaryMergeMsg(req.WtBranch),
		tagNextRootBumpApplyStdout(short),
		donePushConfirmLine(),
	)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))

	assertLocalTagAtMainHEAD(t, req.MainRepo, "v0.0.2")
	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, "v0.0.2") {
		t.Fatal("v0.0.2 should exist on bare origin after --tag-next --push")
	}
	assertLastEventCommandDone(t, req.WrkHome)
}
```
