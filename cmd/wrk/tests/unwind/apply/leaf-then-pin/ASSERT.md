## Expected

- Exit code 0.
- Leaf main advanced with feature content; local tag `v0.0.2` at leaf main HEAD.
- Bare origin `main` matches leaf main HEAD; origin has `refs/tags/v0.0.2`.
- **Pin-when-primary-is-main:** RepoDir == MainRepo; root Path `go.mod` requires
  `example.com/dot-pkgs v0.0.2` with **no** replace (main is in scope → pin edits it).
- Apply peel banner uses the same **relative display path** as dry-run for the
  nested leaf checkout, e.g.:
  `==== unwind: peel external/dot-pkgs-main-2026-06-30 ====`
  (not bare MainRepo basename alone).

## Side Effects

- Leaf land merges branch into leaf main; tag-next + push publish tip + tag.
- Pin edits root Path module require (here Path == MainRepo); tidy may touch go.sum.
- Leaf external path may be removed by `--done` (also covered by done-removes leaf).

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
			t.Fatalf("apply not implemented yet: exit=%d stderr=%q stdout=%q",
				resp.ExitCode, resp.Stderr, resp.Stdout)
		}
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	// Banner display path for nested leaf external (statusDirLine vs RepoDir).
	display := peelDisplay(t, req, req.DepsLinkedWtDir)
	banner := applyBannerLine(display)
	out := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(out, banner) {
		t.Fatalf("apply banner missing %q\ncombined:\n%s", banner, out)
	}
	if strings.HasPrefix(display, "external/") && !strings.Contains(out, "==== unwind: peel external/") {
		t.Fatalf("apply banner must use external/ relative display; got:\n%s", out)
	}
	// Must not satisfy only with bare MainRepo basename peel banner.
	if strings.Contains(out, applyBannerLine(labelDotPkgs)) && !strings.Contains(out, banner) {
		t.Fatalf("banner must not be bare basename only; want %q\ngot:\n%s", banner, out)
	}
	// pin-when-primary-is-main: primary Path is MainRepo (in scope).
	if req.RepoDir == "" || req.MainRepo == "" {
		t.Fatal("RepoDir and MainRepo required")
	}
	if resolvePath(t, req.RepoDir) != resolvePath(t, req.MainRepo) {
		t.Fatalf("leaf-then-pin expects primary Path == MainRepo; RepoDir=%s MainRepo=%s",
			req.RepoDir, req.MainRepo)
	}
	assertLeafMainAdvancedAndTagged(t, req)
	assertConsumerPinned(t, req)
}
```
