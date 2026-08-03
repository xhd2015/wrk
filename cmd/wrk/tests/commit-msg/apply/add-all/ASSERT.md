## Expected

- Exit code 0.
- HEAD subject is exactly `feat: add all`.
- `change.go` is tracked in HEAD (not untracked).

## Side Effects

- `--add-all` staged untracked work then committed.

## Exit Code

- 0

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != "feat: add all" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: add all", resp.Stdout, resp.Stderr)
	}

	// change.go should exist in the tree at HEAD
	if _, err := os.Stat(filepath.Join(req.RepoDir, "change.go")); err != nil {
		t.Fatalf("change.go missing after --add-all commit: %v", err)
	}
	cmd := exec.Command("git", "ls-files", "--error-unmatch", "change.go")
	cmd.Dir = req.RepoDir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("change.go should be tracked after --add-all commit: %v\n%s", err, out)
	}
	// Working tree clean preferred
	st := exec.Command("git", "status", "--porcelain")
	st.Dir = req.RepoDir
	st.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	out, err := st.CombinedOutput()
	if err != nil {
		t.Fatalf("git status: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "change.go") {
		t.Fatalf("change.go still dirty/untracked after --add-all commit; status:\n%s", out)
	}
}
```
