## Expected Output

Primary merge message, blank line, then stable push confirmation:

```
merged branch <WtBranch> into main

pushed main → origin/main
```

## Expected

- Exit code 0.
- Stdout: primary `merged branch <WtBranch> into main`, blank line, `pushed main → origin/main`.
- Stderr empty.
- Worktree directory gone; branch deleted; main has `feature-work`.
- Bare origin `refs/heads/main` equals post-merge main HEAD (branch tip pushed).
- No requirement that any tags were created or pushed (branch-only P2).

## Side Effects

- `origin/main` advances to include the merge-back commit.
- Linked worktree and its branch removed as for plain `--done`.

## Exit Code

- 0

```go
import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	primary := fmt.Sprintf("merged branch %s into main", req.WtBranch)
	want := primaryThenPushStdout(primary, donePushConfirmLine())
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))

	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set by setupDonePushWithOrigin")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
}
```
