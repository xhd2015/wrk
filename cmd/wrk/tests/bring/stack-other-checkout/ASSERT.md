## Expected

- Exit 0.
- Stdout is the new external abs path under primary/`external/`.
- Primary go.mod has `replace example.com/dep => <external>`.
- Other stack checkout (`external/kool`) also has absolute replace to the same external path.
- No `SKIP local dep replacement` on stderr.
- Bring stays quiet (no `==== dep-replace ====` tree).

## Exit Code

- 0

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)

	mod, err := readBringGoMod(req.ConsumerModDir)
	if err != nil {
		t.Fatalf("read primary go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("primary missing replace %s => %s: %+v", bringDepModulePath, wantPath, mod.Replace)
	}

	koolMod, err := readBringGoMod(req.ConsumerModDir2)
	if err != nil {
		t.Fatalf("read kool go.mod: %v", err)
	}
	if !bringHasReplaceForModule(koolMod, bringDepModulePath, wantPath) {
		t.Fatalf("kool missing replace %s => %s: %+v", bringDepModulePath, wantPath, koolMod.Replace)
	}

	// kool filesystem replace on primary retained
	koolPath := filepath.Join(req.ConsumerTop, "external", "kool")
	if !bringHasReplaceForModule(mod, "example.com/kool", koolPath) &&
		!bringHasAnyReplaceForModule(mod, "example.com/kool") {
		t.Fatalf("primary should retain kool replace; got %+v", mod.Replace)
	}

	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
	assertNotContains(t, resp.Stdout, "==== dep-replace")
}
```
