## Expected

- Exit 0.
- Stdout is the new external abs path.
- Consumer go.mod has absolute replace to the external dep root.
- Brought `cmd/go.mod` still has `replace example.com/dep => ../` (not absolute).
- Bring stays quiet (no dep-replace tree).

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

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
		t.Fatalf("read consumer go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("consumer missing replace %s => %s: %+v", bringDepModulePath, wantPath, mod.Replace)
	}

	cmdGoMod := filepath.Join(wantPath, "cmd", "go.mod")
	data, err := os.ReadFile(cmdGoMod)
	if err != nil {
		t.Fatalf("read brought cmd go.mod: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "replace "+bringDepModulePath+" => ../") {
		t.Fatalf("brought cmd/go.mod should keep relative replace; got:\n%s", body)
	}
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace ") && strings.Contains(trim, bringDepModulePath) {
			parts := strings.Split(trim, "=>")
			if len(parts) == 2 && filepath.IsAbs(strings.TrimSpace(parts[1])) {
				t.Fatalf("brought cmd replace must not become absolute:\n%s", body)
			}
		}
	}

	assertNotContains(t, resp.Stdout, "==== dep-replace")
	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
}
```
