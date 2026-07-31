## Expected

- Non-zero exit code.
- Stderr contains `unexpected arguments` (preferred: `wrk: unexpected arguments`).
- Must **not** treat the second path as a successful second bring or as consumer workDir.
- Prefer no `external/` created (reject early).

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --bring with extra positional, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "unexpected arguments")

	// Must not have brought both as if multi-value sugar worked.
	want2 := bringExternalWorktreePath(req.ConsumerTop, "mydep2", "main", 0)
	assertFileNotExists(t, want2)
	// Prefer reject before materializing any external/.
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```
