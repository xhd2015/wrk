## Expected

- Exit code 0.
- Worktree exists exactly at `{WorkRoot}/dst` (no default-naming suffix).
- External for mydep1 is under `dst/external/`; stdout includes `dst` and that external path.
- Source `src/external` does not exist; `src/go.mod` unchanged (bring writes replace only in the new WT).
- No worktree under `{WRK_HOME}/worktrees`.

## Side Effects

- Relocated contract: `<target-dir>` + `--bring` is **allowed** (create then bring into the spawn path).

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	dst := req.SpawnDir
	ext1 := createBringExternalPath(dst, "mydep1")
	assertFileExists(t, dst)
	assertGitFileIsWorktreeLink(t, dst)
	assertWorktreeListContains(t, req.MainRepo, dst)
	if !createBringStdoutHasLine(resp.Stdout, dst) {
		t.Fatalf("stdout should include spawn path %q; got %q", dst, resp.Stdout)
	}
	assertFileExists(t, ext1)
	if !createBringStdoutHasLine(resp.Stdout, ext1) {
		t.Fatalf("stdout should include external %q; got %q", ext1, resp.Stdout)
	}

	assertFileNotExists(t, filepath.Join(req.MainRepo, "external"))
	createBringAssertGoModUnchanged(t, req, req.MainRepo)
	assertFileNotExists(t, worktreePath(req.WrkHome, createBringSrcName, "main", wrkDate, 0))

	mod, err := readCreateBringGoMod(dst)
	if err != nil {
		t.Fatalf("read spawn go.mod: %v", err)
	}
	if !createBringHasReplace(mod, createBringDep1Module, ext1) {
		t.Fatalf("spawn go.mod missing replace %s => %s", createBringDep1Module, ext1)
	}
	ok, err := createBringGitignoreHasExternal(dst)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf("spawn .gitignore should contain /external")
	}
}
```
