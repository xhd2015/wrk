## Expected

- Exit code 0 (RED while apply stubbed).
- No land required / no worktree removed (never had linked WT).
- Local tag `v0.0.2` at main HEAD; bare origin main == local main; origin has tag.
- Must not error for missing `--done`/`--merge-back` (already main, no edges).

## Side Effects

- tag-next creates next root tag; push publishes branch + tag to local bare remote.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		combined := resp.Stdout + "\n" + resp.Stderr
		if strings.Contains(combined, "not implemented") {
			t.Fatalf("apply not implemented yet (expected RED until P4 lands): exit=%d stderr=%q stdout=%q",
				resp.ExitCode, resp.Stderr, resp.Stdout)
		}
		// Must not be a land-flag validation error for already-main.
		if strings.Contains(combined, "merge-back") || strings.Contains(combined, "--done") {
			t.Fatalf("already-main peel must not require land flags; stderr=%q stdout=%q",
				resp.Stderr, resp.Stdout)
		}
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertLocalTagAtMainHEAD(t, req.MainRepo, req.ExpectedPinVersion)
	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, req.ExpectedPinVersion) {
		t.Fatalf("%s should exist on bare origin after --tag-next --push", req.ExpectedPinVersion)
	}
}
```
