## Expected

- Exit code 0.
- Paths: `./external/mydep1-main-{date}` and `./external/mydep2-main-{date}` (no `-1`; basenames differ).
- Branch in mydep1 is `main-{date}` (not `mydep1-main-{date}`).
- Branch in mydep2 is `main-{date}` (not `mydep2-main-{date}`).
- Each worktree checks out its own dep's `main-{date}` branch.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	assertWorktreeListContains(t, allDepsDepMainRepo(dep1), wantDep1)
	assertWorktreeListContains(t, allDepsDepMainRepo(dep2), wantDep2)

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, dep1, wantBranch)
	assertBranchExists(t, dep2, wantBranch)
	assertBranchNotExists(t, dep1, "mydep1-"+wantBranch)
	assertBranchNotExists(t, dep2, "mydep2-"+wantBranch)
	assertBranchCheckedOutInWorktree(t, wantDep1, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantDep2, wantBranch)
}
```
