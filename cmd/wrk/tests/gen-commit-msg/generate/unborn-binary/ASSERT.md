## Expected

- Exit code 0.
- Stderr does not contain `could not resolve 'HEAD'` / `git restore --staged failed`.
- Binary is unstaged; `app.go` remains staged; binary still on disk.
- Stdout contains mock title `feat: add feature`.
- HEAD stays unborn (no `--commit`).

## Side Effects

- Auto-unstage mutates the index (binary removed from staged set).

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
	assertExitZero(t, resp)

	if strings.Contains(resp.Stderr, "could not resolve") ||
		strings.Contains(resp.Stderr, "git restore --staged failed") {
		t.Fatalf("unborn HEAD must not fail auto-unstage RestoreStaged, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "feat: add feature") {
		t.Fatalf("stdout missing title, got:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}

	binRel := req.BinaryRel
	if binRel == "" {
		binRel = "blob.bin"
	}
	staged := gitStagedNames(t, req.RepoDir)
	joined := strings.Join(staged, "\n")
	if strings.Contains(joined, binRel) {
		t.Fatalf("binary %q must be unstaged after generate, staged=%v", binRel, staged)
	}
	if !strings.Contains(joined, "app.go") {
		t.Fatalf("text file app.go must remain staged, staged=%v", staged)
	}
	if _, statErr := os.Stat(filepath.Join(req.RepoDir, binRel)); statErr != nil {
		t.Fatalf("binary %q must remain on disk: %v", binRel, statErr)
	}
	if gitHEADExists(t, req.RepoDir) {
		t.Fatalf("HEAD must stay unborn without --commit")
	}
}
```
