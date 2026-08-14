## Expected

- Exit code 0.
- Create path includes slug `fix-login` after the date; branch does too.
- External for mydep1 lives under that slug WT.
- Last event `command=="create"`; `args` include `-t` and `--bring`.

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

	wt := createBringDefaultWTWithTask(req, "fix login")
	ext1 := createBringExternalPath(wt, "mydep1")
	slug := slugify("fix login")
	if filepath.Base(wt) == "" || !createBringStdoutHasLine(resp.Stdout, wt) {
		t.Fatalf("stdout should include slugged create path %q; got %q", wt, resp.Stdout)
	}
	assertContains(t, filepath.Base(wt), slug)
	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	wantBranch := branchNameWithTask("main", wrkDate, slug, 0)
	assertBranchExists(t, req.MainRepo, wantBranch)
	assertBranchCheckedOutInWorktree(t, wt, wantBranch)

	assertFileExists(t, ext1)
	if !createBringStdoutHasLine(resp.Stdout, ext1) {
		t.Fatalf("stdout should include external %q; got %q", ext1, resp.Stdout)
	}
	mod, err := readCreateBringGoMod(wt)
	if err != nil {
		t.Fatalf("read new WT go.mod: %v", err)
	}
	if !createBringHasReplace(mod, createBringDep1Module, ext1) {
		t.Fatalf("new WT go.mod missing replace %s => %s", createBringDep1Module, ext1)
	}

	ev := createBringLastEvent(t, req.WrkHome)
	if ev.Command != "create" {
		t.Fatalf("event command: want %q, got %q", "create", ev.Command)
	}
	if !createBringArgsContain(ev.Args, "-t") {
		t.Fatalf("event args should include -t, got %v", ev.Args)
	}
	if !createBringArgsContain(ev.Args, "--bring") {
		t.Fatalf("event args should include --bring, got %v", ev.Args)
	}
	if !createBringArgsContain(ev.Args, req.DepPath) {
		t.Fatalf("event args should include dep path %q, got %v", req.DepPath, ev.Args)
	}
}
```
