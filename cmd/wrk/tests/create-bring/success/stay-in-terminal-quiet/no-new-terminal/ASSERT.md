## Expected

- Exit code 0.
- Stdout is exactly the create worktree path (no external path lines).
- Stderr has no `will bring:` and no `SKIP local dep replacement`.
- Both externals exist under the new WT; replaces applied.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
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

	if strings.TrimSpace(resp.Stdout) != wt {
		t.Fatalf("stdout should be create path only %q; got %q", wt, resp.Stdout)
	}
	if createBringStdoutHasLine(resp.Stdout, ext1) || createBringStdoutHasLine(resp.Stdout, ext2) {
		t.Fatalf("stdout must not include external paths under --no-new-terminal; got %q", resp.Stdout)
	}
	assertNotContains(t, resp.Stderr, "will bring:")
	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")

	assertFileExists(t, wt)
	assertFileExists(t, ext1)
	assertFileExists(t, ext2)
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
}
```