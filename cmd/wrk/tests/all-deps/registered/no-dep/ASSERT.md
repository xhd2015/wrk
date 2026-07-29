## Expected Output

```
wrk example.com/dep1 at ./external/mydep1-main-2026-06-30
wrk example.com/dep2 at ./external/mydep2-main-2026-06-30
wrked 2 deps
```

## Expected

- Exit code 0.
- Stdout lists both deps (project-path order) then `wrked 2 deps` (same as basic, not `would:`).
- Both external paths exist as linked worktrees.
- Consumer `go.mod` has **no** replace for either dep (byte-identical to snapshot).
- Consumer `.gitignore` contains `/external` once.

## Exit Code

- 0

```go
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
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

	before, err := os.ReadFile(filepath.Join(req.WorkRoot, "go.mod.before"))
	if err != nil {
		t.Fatalf("read go.mod.before: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(req.RepoDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("go.mod changed under --all-deps --no-dep\nbefore:\n%s\nafter:\n%s", before, after)
	}

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep1", "") {
		t.Fatalf("go.mod should have no replace for example.com/dep1 under --no-dep")
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep2", "") {
		t.Fatalf("go.mod should have no replace for example.com/dep2 under --no-dep")
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
