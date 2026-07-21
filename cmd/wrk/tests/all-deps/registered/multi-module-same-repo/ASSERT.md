## Expected

- Exit code 0.
- Stdout lists dep1 at `svc-a` then dep2 at `svc-b` (mod/scan Dir order within the project), then `wrked 2 deps`.
- Exactly one external worktree for `myrepo`.

## Exit Code

- 0

```go
import (
	"os"

	"github.com/xhd2015/doctest/assert"
	"fmt"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	myrepo := allDepsDepDir(req.WorkRoot, "myrepo")
	wantRel1 := nestedExternalRelSubPath("myrepo", "svc-a")
	wantRel2 := nestedExternalRelSubPath("myrepo", "svc-b")
	wantStdout := fmt.Sprintf("wrk example.com/dep1 at %s\nwrk example.com/dep2 at %s\nwrked 2 deps\n", wantRel1, wantRel2)
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	repoExternal := allDepsExternalAbsPath(req.ConsumerTop, "myrepo")
	assertFileExists(t, repoExternal)
	assertGitFileIsWorktreeLink(t, repoExternal)
	assertWorktreeListContains(t, allDepsDepMainRepo(myrepo), repoExternal)
	assertWorktreeListNotContains(t, req.ConsumerTop, repoExternal)
	suffixed := repoExternal + "-1"
	assertWorktreeListNotContains(t, req.ConsumerTop, suffixed)
	if _, err := os.Stat(suffixed); err == nil {
		t.Fatalf("no suffixed second worktree should exist at %s", suffixed)
	}

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	wantSub1 := nestedExternalAbsSubPath(req.ConsumerTop, "myrepo", "svc-a")
	wantSub2 := nestedExternalAbsSubPath(req.ConsumerTop, "myrepo", "svc-b")
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantSub1) {
		t.Fatalf("go.mod missing replace example.com/dep1 => %s: %+v", wantSub1, mod.Replace)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantSub2) {
		t.Fatalf("go.mod missing replace example.com/dep2 => %s: %+v", wantSub2, mod.Replace)
	}
}
```