## Expected Output

```
merged branch <WtBranch> into main

<flushed tag-next body>
<flushed sync body>
```

## Expected

- Exit code 0.
- Concurrent ship: stderr has `] ship · concurrent` and `] merge-back`.
- Stdout: primary + flushed `tagged` + `synced:`; no `pushed`.
- wtA **kept**; wtB HEAD == main HEAD; local `v0.0.2` at main HEAD.
- Event command `"merge-back"`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertShipStderrMarkers(t, resp.Stderr, true)
	if !strings.Contains(resp.Stderr, "] merge-back") {
		t.Fatalf("expected merge-back stage marker; stderr=%q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "worktree removed:")

	primary := strings.TrimSuffix(primaryMergeMsg(req.WtBranch), "\n")
	assertContains(t, resp.Stdout, primary)
	assertContains(t, resp.Stdout, "tagged")
	assertContains(t, resp.Stdout, "synced:")
	if strings.Contains(resp.Stdout, "pushed") {
		t.Fatalf("stdout must not include pushed without --push; got %q", resp.Stdout)
	}

	assertSourceWorktreeKept(t, req)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertHEAD(t, req.Wt2Dir, revParseHEAD(t, req.MainRepo))

	assertLocalTagAtMainHEAD(t, req.MainRepo, "v0.0.2")
	assertLastEventCommandMergeBack(t, req.WrkHome)
}
```
