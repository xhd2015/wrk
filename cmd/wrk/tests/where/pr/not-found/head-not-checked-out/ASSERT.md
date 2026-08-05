## Expected

- Non-zero exit.
- Stdout empty.
- Stderr names the PR (number `42` and/or URL), head branch `feature-pr`, and
  repo `acme/app` (or equivalent). Soft: at least two of {number/url, branch, repo}.

## Errors

- Local main matches origin but no live worktree has head checked out.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when head not checked out; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := resp.Stderr
	hits := 0
	if strings.Contains(se, "42") || strings.Contains(se, wherePrURL) ||
		strings.Contains(strings.ToLower(se), "pull request") ||
		strings.Contains(strings.ToLower(se), " pr ") {
		hits++
	}
	if strings.Contains(se, wherePrHeadBranch) || strings.Contains(se, "feature-pr") {
		hits++
	}
	if strings.Contains(se, "acme/app") || strings.Contains(se, wherePrRepo) {
		hits++
	}
	if hits < 2 {
		t.Fatalf("stderr should name PR + head branch + repo (need ≥2); got %q", se)
	}
}
```
