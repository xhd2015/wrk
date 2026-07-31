## Expected

- Exit code 0.
- Not mutually exclusive (`gen-commit-msg` + `push` + `pr` compose).
- HEAD subject on worktree is exactly `feat: compose pr` (gen-commit stage committed).
- Origin `feature-pr` equals post-commit local HEAD (push stage ran after commit).
- Stdout order: commit title appears before push confirm; push confirm before `PR created`.
- Stdout includes push confirm and new-PR success tokens + URL.
- Fake `gh`: create only (body = comment); no issue comment.
- Stderr may include gen-commit logs (`Running git commit...`, generated-message banners); must not contain `mutually exclusive`.

## Side Effects

- New commit on feature branch; tip pushed; PR created with body = `--comment`.

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
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("gen-commit + push + pr still mutually exclusive; stderr=%q", resp.Stderr)
	}

	subject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	if subject != prComposeCommitTitle {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q",
			subject, prComposeCommitTitle, resp.Stdout, resp.Stderr)
	}

	branch := req.WtBranch
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)

	out := resp.Stdout
	if !strings.Contains(out, prComposeCommitTitle) {
		t.Fatalf("stdout missing commit title %q; stdout=%q stderr=%q", prComposeCommitTitle, out, resp.Stderr)
	}
	pushLine := strings.TrimSuffix(prPushConfirmLine(branch), "\n")
	if !strings.Contains(out, pushLine) {
		t.Fatalf("stdout missing push confirm %q; stdout=%q", pushLine, out)
	}
	if !strings.Contains(out, "PR created") {
		t.Fatalf("stdout missing PR created; stdout=%q", out)
	}
	if !strings.Contains(out, prDefaultURL) {
		t.Fatalf("stdout missing PR URL; stdout=%q", out)
	}

	idxTitle := strings.Index(out, prComposeCommitTitle)
	idxPush := strings.Index(out, pushLine)
	idxPR := strings.Index(out, "PR created")
	if idxTitle < 0 || idxPush < 0 || idxPR < 0 {
		t.Fatalf("missing ordered tokens; title=%d push=%d pr=%d\n%s", idxTitle, idxPush, idxPR, out)
	}
	if !(idxTitle < idxPush && idxPush < idxPR) {
		t.Fatalf("want gen-commit title before push before PR created; title=%d push=%d pr=%d\n%s",
			idxTitle, idxPush, idxPR, out)
	}

	// Staged file should be in the new commit (not left staged only).
	cached := gitOutputIsolated(t, req.WtDir, "diff", "--cached", "--name-only")
	if strings.Contains(cached, "compose-stage.go") {
		t.Fatalf("compose-stage.go should be committed, still staged; cached=%q", cached)
	}
	names := gitOutputIsolated(t, req.WtDir, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(names, "compose-stage.go") {
		t.Fatalf("HEAD commit should include compose-stage.go; show=%q", names)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
