## Expected Output

```
merged branch <WtBranch> into main

<flushed push body>
<flushed sync body>
```

## Expected

- Exit code 0.
- Concurrent ship: stderr has `] ship · concurrent`, `] done`, progress label `push`; must not show `tag-next+push`.
- Stdout: primary + `pushed` + `synced:`; no `tagged`.
- wtA removed; wtB HEAD == main; `feature-work` on main.
- No local `v0.0.2`; origin/main == main HEAD.
- Event command `"done"`.

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
	if !strings.Contains(resp.Stderr, "] done") {
		t.Fatalf("expected done stage marker; stderr=%q", resp.Stderr)
	}
	assertContains(t, resp.Stderr, "push")
	assertNotContains(t, resp.Stderr, "tag-next+push")
	assertNotContains(t, resp.Stderr, "tag-next")

	primary := strings.TrimSuffix(primaryMergeMsg(req.WtBranch), "\n")
	assertContains(t, resp.Stdout, primary)
	assertContains(t, resp.Stdout, "pushed")
	assertContains(t, resp.Stdout, "synced:")
	if strings.Contains(resp.Stdout, "tagged") {
		t.Fatalf("stdout must not include tagged without --tag-next; got %q", resp.Stdout)
	}

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertHEAD(t, req.Wt2Dir, revParseHEAD(t, req.MainRepo))

	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not exist without --tag-next")
	}
	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	assertLastEventCommandDone(t, req.WrkHome)
}
```