## Expected

- Exit code 0.
- Create path `{WRK_HOME}/worktrees/src-main-2026-06-30` exists as a linked worktree and is a stdout line.
- External paths for mydep1 and mydep2 under **that** worktree are stdout lines.
- `src/external` does not exist; `src/go.mod` is unchanged (no replace).
- New WT `go.mod` has replaces for both modules; `/external` gitignore on the new WT.
- Last event `command=="create"`; `args` contain `--bring` and **each** dep path.

## Side Effects

- New worktree under `{WRK_HOME}/worktrees`; bring consumer is that path, not `src`.

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

	wt := createBringDefaultWT(req)
	ext1 := createBringExternalPath(wt, "mydep1")
	ext2 := createBringExternalPath(wt, "mydep2")
	req.ConsumerTop = wt
	req.ExternalWtDir = ext1
	req.ExternalWtDir2 = ext2

	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	assertWorktreeListContains(t, req.MainRepo, wt)
	if !createBringStdoutHasLine(resp.Stdout, wt) {
		t.Fatalf("stdout should include create path %q as a line; got %q", wt, resp.Stdout)
	}
	if !createBringStdoutHasLine(resp.Stdout, ext1) {
		t.Fatalf("stdout should include external %q as a line; got %q", ext1, resp.Stdout)
	}
	if !createBringStdoutHasLine(resp.Stdout, ext2) {
		t.Fatalf("stdout should include external %q as a line; got %q", ext2, resp.Stdout)
	}

	assertFileExists(t, ext1)
	assertFileExists(t, ext2)
	assertGitFileIsWorktreeLink(t, ext1)
	assertGitFileIsWorktreeLink(t, ext2)
	assertWorktreeListContains(t, req.DepPath, ext1)
	assertWorktreeListContains(t, req.SecondRepo, ext2)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "external"))
	createBringAssertGoModUnchanged(t, req, req.MainRepo)

	mod, err := readCreateBringGoMod(wt)
	if err != nil {
		t.Fatalf("read new WT go.mod: %v", err)
	}
	if !createBringHasReplace(mod, createBringDep1Module, ext1) {
		t.Fatalf("new WT go.mod missing replace %s => %s: %+v", createBringDep1Module, ext1, mod.Replace)
	}
	if !createBringHasReplace(mod, createBringDep2Module, ext2) {
		t.Fatalf("new WT go.mod missing replace %s => %s: %+v", createBringDep2Module, ext2, mod.Replace)
	}
	ok, err := createBringGitignoreHasExternal(wt)
	if err != nil {
		t.Fatalf("read new WT .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf("new WT .gitignore should contain /external")
	}
	assertContains(t, resp.Stderr, "will bring:")

	ev := createBringLastEvent(t, req.WrkHome)
	if ev.Command != "create" {
		t.Fatalf("event command: want %q, got %q args=%v", "create", ev.Command, ev.Args)
	}
	if !createBringArgsContain(ev.Args, "--bring") {
		t.Fatalf("event args should include --bring, got %v", ev.Args)
	}
	if !createBringArgsContain(ev.Args, req.DepPath) {
		t.Fatalf("event args should include first dep %q, got %v", req.DepPath, ev.Args)
	}
	if !createBringArgsContain(ev.Args, req.SecondRepo) {
		t.Fatalf("event args should include second dep %q, got %v", req.SecondRepo, ev.Args)
	}
}
```
