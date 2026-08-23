## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep`.
- `go-pkgs/go.mod` has `replace example.com/dep => <external path>`.
- `tools/go.mod` has NO replace for dep.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

import "path/filepath"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	// go-pkgs/ must have replace.
	mod, err := readBringGoMod(req.ConsumerModDir)
	if err != nil {
		t.Fatalf("read go-pkgs/go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("go-pkgs/go.mod missing replace %s => %s: %+v", bringDepModulePath, wantPath, mod.Replace)
	}

	// tools/ must NOT have the replace.
	toolsDir := filepath.Join(req.ConsumerTop, "tools")
	toolsMod, err := readBringGoMod(toolsDir)
	if err != nil {
		t.Fatalf("read tools/go.mod: %v", err)
	}
	if bringHasReplaceForModule(toolsMod, bringDepModulePath, wantPath) {
		t.Fatalf("tools/go.mod should NOT have replace for %s: %+v", bringDepModulePath, toolsMod.Replace)
	}
}
```