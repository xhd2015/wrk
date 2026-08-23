## Expected Output

```text
.
external/mydep
```

## Expected

- Exit code 0; stderr empty.
- Stdout lists relative repo paths from checkout root: `.` and the external dep path
  (`filepath.ToSlash(Rel(consumer, external))`).
- Same discovery completeness as `--status` (incomplete warm index must not hide external).

## Exit Code

- 0

```go
import (
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
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}

	extRel, err := filepath.Rel(req.RepoDir, req.ExternalWtDir)
	if err != nil {
		t.Fatalf("rel external: %v", err)
	}
	extRel = filepath.ToSlash(extRel)

	// Membership + count (order: path-sorted discovery typically . then external/…).
	lines := strings.Split(strings.TrimSuffix(resp.Stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 repo lines, got %d (%q):\n%s", len(lines), lines, resp.Stdout)
	}
	seen := map[string]bool{}
	for _, ln := range lines {
		seen[ln] = true
	}
	if !seen["."] {
		t.Fatalf("repos missing %q, got:\n%s", ".", resp.Stdout)
	}
	if !seen[extRel] {
		t.Fatalf("repos missing %q, got:\n%s", extRel, resp.Stdout)
	}

	// Exact path-sorted body when product keeps path sort (consumer < external).
	want := ".\n" + extRel + "\n"
	if resp.Stdout != want {
		t.Fatalf("repos stdout mismatch:\nwant %q\ngot  %q", want, resp.Stdout)
	}
}
```
