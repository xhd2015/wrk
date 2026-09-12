## Expected Output

```
merged branch <WtBranch> into main

<flushed tag-next+push body>
<flushed sync body>
```

## Expected

- Exit code 0.
- Concurrent ship: stderr has `] ship · concurrent` and `] merge-back`; progress labels on stderr.
- Stdout: primary + flushed `tagged` / `pushed` / `synced:`.
- wtA **remains**; wtB HEAD == main; `feature-work` on main.
- Local + origin `v0.0.2`; origin/main == main HEAD.
- Event command `"merge-back"`.

## Side Effects

- Remaining worktree synced; tags created; branch+tags pushed; source worktree kept.

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
	assertContains(t, resp.Stdout, "pushed")
	assertContains(t, resp.Stdout, "synced:")

	assertSourceWorktreeKept(t, req)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertHEAD(t, req.Wt2Dir, revParseHEAD(t, req.MainRepo))

	assertLocalTagAtMainHEAD(t, req.MainRepo, "v0.0.2")
	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, "v0.0.2") {
		t.Fatal("v0.0.2 should exist on bare origin after full pipeline")
	}
	assertLastEventCommandMergeBack(t, req.WrkHome)
}
```
