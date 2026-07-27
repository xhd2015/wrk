## Expected

- Exit code 0.
- Stdout lists dep1 then dep2 (**lexicographic project path order**: `.../mydep1` before `.../mydep2`), each at `./external/mydepN-main-2026-06-30`, then `wrked 2 deps`.
- Both external paths exist as linked git worktrees owned by their dep repos.
- Consumer `go.mod` has a `replace` for each dep at its external path.
- Consumer `.gitignore` contains `/external` exactly once.

## Exit Code

- 0

```go
import (

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"

	"fmt"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	wantDep1 := allDepsExternalAbsPath(req.ConsumerTop, "mydep1")
	wantDep2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	wantStdout := fmt.Sprintf("wrk example.com/dep1 at %s\nwrk example.com/dep2 at %s\nwrked 2 deps\n",
		allDepsExternalRelPath("mydep1"), allDepsExternalRelPath("mydep2"))
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	assertFileExists(t, wantDep1)
	assertFileExists(t, wantDep2)
	assertGitFileIsWorktreeLink(t, wantDep1)
	assertGitFileIsWorktreeLink(t, wantDep2)
	assertWorktreeListContains(t, allDepsDepMainRepo(dep1), wantDep1)
	assertWorktreeListContains(t, allDepsDepMainRepo(dep2), wantDep2)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantDep1)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantDep2)

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantDep1) {
		t.Fatalf("go.mod missing replace example.com/dep1 => %s: %+v", wantDep1, mod.Replace)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantDep2) {
		t.Fatalf("go.mod missing replace example.com/dep2 => %s: %+v", wantDep2, mod.Replace)
	}

	n, err := allDepsCountGitignoreExternalLines(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if n != 1 {
		t.Fatalf(".gitignore should contain /external exactly once, got %d", n)
	}
}
```