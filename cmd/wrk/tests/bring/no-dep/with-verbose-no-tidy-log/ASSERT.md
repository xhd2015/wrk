## Expected

- Exit code 0.
- External worktree created; go.mod unchanged (no replace).
- Stderr must **not** contain `mod tidy` or a `$ go` tidy pre-line (`$ go` + `mod tidy`).
- Stderr may contain git worktree logging (optional; not required for this leaf).

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

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertBringGoModUnchanged(t, req, req.RepoDir)

	assertNotContains(t, resp.Stderr, "mod tidy")
	// Guard against any verbose go pre-line form for tidy.
	if strings.Contains(resp.Stderr, "$ go") && strings.Contains(resp.Stderr, "tidy") {
		t.Fatalf("stderr must not log go mod tidy under --no-dep -v, got %q", resp.Stderr)
	}
}
```
