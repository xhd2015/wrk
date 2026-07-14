## Expected

- Exit code 0.
- Consumer worktree renamed to new slug path.
- `go.mod` replace for `example.com/dep` points at the **new** absolute external
  path under the renamed consumer wt (not the pre-rename consumer path).
- External dep worktree still exists at the new path.

## Exit Code

- Zero

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	newSlug := slugify("new slug")
	newConsumerPath := worktreePathWithTask(req.WrkHome, "consumer", "main", wrkDate, newSlug, 0)
	assertFileExists(t, newConsumerPath)

	oldExtBase := filepath.Base(req.ExternalWtDir)
	wantReplace := filepath.Join(newConsumerPath, "external", oldExtBase)
	wantExtPath := wantReplace

	mod, err := propagateReadGoMod(newConsumerPath)
	if err != nil {
		t.Fatalf("read go.mod after set-task: %v", err)
	}
	gotReplace := propagateReplacePathForModule(mod, depModulePath)
	if gotReplace != wantReplace {
		t.Fatalf("go.mod replace for %s = %q, want %q (abs path must follow consumer rename)", depModulePath, gotReplace, wantReplace)
	}
	if strings.HasPrefix(gotReplace, req.WtDir) {
		t.Fatalf("go.mod replace still under old consumer path %q: %q", req.WtDir, gotReplace)
	}

	assertFileExists(t, wantExtPath)
}
```