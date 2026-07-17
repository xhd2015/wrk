## Expected

- Exit 0; not mutually exclusive.
- Gen-commit dry plan present.
- Tag-next and push plans present; tag before push.
- No done/merge-back.
- HEAD subject and staged file preserved.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertNoMutexReject(t, resp.Stderr)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	outAll := resp.Stdout + "\n" + resp.Stderr
	hasMockB := strings.Contains(outAll, "would generate commit message")
	hasWouldCommit := strings.Contains(strings.ToLower(resp.Stderr), "would:") &&
		strings.Contains(resp.Stderr, "git commit")
	if !hasMockB && !hasWouldCommit {
		t.Fatalf("expected gen-commit dry plan; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "1 tag planned") {
		t.Fatalf("expected tag-next plan; stdout=%q", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "would: git push origin main") {
		t.Fatalf("expected push plan; stdout=%q", resp.Stdout)
	}
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	idxPush := strings.Index(resp.Stdout, "would: git push origin main")
	if idxTag > idxPush {
		t.Fatalf("want tag before push; tag=%d push=%d\n%s", idxTag, idxPush, resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "merge --ff-only")
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run; stderr=%q", resp.Stderr)
	}
	wantSubject := readAPBaseline(t, req, "main.head-subject")
	gotSubject := strings.TrimSpace(gitOutputIsolated(t, req.MainRepo, "log", "-1", "--format=%s"))
	if gotSubject != wantSubject {
		t.Fatalf("main HEAD subject changed: before=%q after=%q", wantSubject, gotSubject)
	}
	cached := gitOutputIsolated(t, req.MainRepo, "diff", "--cached", "--name-only")
	if !strings.Contains(cached, "staged-for-commit.go") {
		t.Fatalf("staged file should remain staged; cached=%q", cached)
	}
	_ = filepath.Separator
}
```
