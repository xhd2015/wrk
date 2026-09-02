## Expected

- Exit code 0; stdout is create path only.
- Stderr has no `will bring:` and no `SKIP local dep replacement`.
- Both externals exist; only dep1 is replaced in the new WT go.mod.

## Exit Code

- 0

```go
import (
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
	if strings.TrimSpace(resp.Stdout) != wt {
		t.Fatalf("stdout should be create path only %q; got %q", wt, resp.Stdout)
	}
	assertNotContains(t, resp.Stderr, "will bring:")
	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")

	assertFileExists(t, ext1)
	assertFileExists(t, ext2)

	mod, err := readCreateBringGoMod(wt)
	if err != nil {
		t.Fatalf("read new WT go.mod: %v", err)
	}
	if !createBringHasReplace(mod, createBringDep1Module, ext1) {
		t.Fatalf("new WT go.mod missing replace %s => %s: %+v", createBringDep1Module, ext1, mod.Replace)
	}
	if createBringHasAnyReplace(mod, createBringDep2Module) {
		t.Fatalf("new WT go.mod should not replace non-required %s: %+v", createBringDep2Module, mod.Replace)
	}
}
```
