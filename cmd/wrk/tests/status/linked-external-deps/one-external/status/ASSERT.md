## Expected Output

```text
Dir:          .
Branch:       …
Commit:       …
Status:       clean
Master:       …

Dir:          external/mydep
Branch:       …
Commit:       …
Status:       clean
Master:       …
```

## Expected

- Exit code 0; stderr empty.
- **Exactly 2** `Dir:` blocks (consumer + one external dep).
- Consumer block is `Dir: .` with `Master:` vs consumer main.
- External block `Dir` is the relative path under consumer (`statusDirLine`);
  includes `Master:` vs **dep** main (other main owns the worktree).
- No `---- external ----` header (linked-cwd scan-only; multi-block OK without header).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
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
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks (consumer + external dep), got %d:\n%s", got, resp.Stdout)
	}
	assertNoExternalSectionHeader(t, resp.Stdout)

	extDirLine := statusDirLine(t, req.RepoDir, req.ExternalWtDir)
	assertStdoutHasDirLine(t, resp.Stdout, ".")
	assertStdoutHasDirLine(t, resp.Stdout, extDirLine)

	// Path-sorted discovery: consumer root then external under it.
	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		linkedScanBlock(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		linkedScanBlock(t, req.RepoDir, req.DepPath, req.ExternalWtDir, req.Wt2Branch, "clean"),
	))
}
```
