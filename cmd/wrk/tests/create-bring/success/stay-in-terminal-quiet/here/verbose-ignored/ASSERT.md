## Expected

- Exit code 0; stdout is create path only.
- Stderr has no timestamped `$ git … worktree add` pre-line and no `Preparing worktree` / `HEAD is now at` stream.
- External still exists under the new WT.

## Exit Code

- 0

```go
import (
	"regexp"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wt := createBringDefaultWT(req)
	ext := createBringExternalPath(wt, "mydep1")
	if strings.TrimSpace(resp.Stdout) != wt {
		t.Fatalf("stdout should be create path only %q; got %q", wt, resp.Stdout)
	}
	if createBringStdoutHasLine(resp.Stdout, ext) {
		t.Fatalf("stdout must not include external path under --here; got %q", resp.Stdout)
	}
	assertFileExists(t, ext)

	reGit := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \$ git `)
	if reGit.MatchString(resp.Stderr) {
		t.Fatalf("--verbose should be ignored with --here; got timestamped git log in stderr %q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "worktree add") {
		t.Fatalf("--verbose should be ignored with --here; stderr has worktree add: %q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Preparing worktree") || strings.Contains(resp.Stderr, "HEAD is now at") {
		t.Fatalf("--verbose should be ignored with --here; stderr has worktree stream: %q", resp.Stderr)
	}
}
```
