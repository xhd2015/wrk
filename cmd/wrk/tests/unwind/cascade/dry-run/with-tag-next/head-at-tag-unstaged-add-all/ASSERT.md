## Expected Output

Phase-1 under `.` includes tip-planned tag-next after gen-commit land plan:

```
[2/4] phase-1 · cross-repo unwind
      .
         would: git add -A
         would: gen-commit-msg
         would: commit
         would: tag-next example.com/root @ v0.0.2
```

(Exact stage banners / indent implementer-owned; asserts lock substrings.)

## Expected

- Exit code 0.
- Stdout contains `would: git add -A` (peeled `--add-all` honored in plan).
- Stdout contains `would: tag-next example.com/root @ v0.0.2` (tip-aware NextTag).
- Order: `git add -A` / gen-commit / commit before `tag-next`.
- Zero mutations (HEAD still at `v0.0.1`; root.go still dirty).

## Side Effects

- None (dry-run).

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
	assertExitZero(t, resp)

	out := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(out, "would: git add -A") {
		t.Fatalf("peeled --add-all must plan git add -A\ncombined:\n%s", out)
	}
	wantTag := "would: tag-next " + unwindRootModule + " @ " + unwindApplyNextTag
	if !strings.Contains(out, wantTag) {
		t.Fatalf("HEAD@LatestTag + unstaged owned + peeled --add-all must tip-plan tag-next\nwant %q\ncombined:\n%s",
			wantTag, out)
	}
	assertContainsInOrder(t, out,
		"would: git add -A",
		"would: gen-commit-msg",
		"would: commit",
		wantTag,
	)
	// Plan-only: HEAD stays at LatestTag; unstaged owned WIP preserved.
	// (Do not use assertUnwindZeroMutations — it requires a DIRTY marker file.)
	gotHEAD := revParseHEAD(t, req.MainRepo)
	if want := readBaselineSHA(t, req, "main.sha"); gotHEAD != want {
		t.Fatalf("main HEAD mutated: got %s want %s", gotHEAD, want)
	}
	tagSHA := revParseRef(t, req.MainRepo, "refs/tags/"+unwindApplyOldTag)
	if tagSHA != gotHEAD {
		t.Fatalf("dry-run must not advance HEAD past %s; HEAD=%s tag=%s",
			unwindApplyOldTag, gotHEAD, tagSHA)
	}
	if tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("dry-run must not create tag %s", unwindApplyNextTag)
	}
	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain", "--", "root.go")
	if strings.TrimSpace(status) == "" {
		t.Fatal("dry-run must leave unstaged root.go WIP")
	}
}
```
