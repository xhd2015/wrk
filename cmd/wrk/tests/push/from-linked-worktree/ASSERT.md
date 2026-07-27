
## Expected Output

```
pushed feature-push → origin/feature-push
```

## Expected

- Exit code 0.
- Stdout: confirm line for the **worktree branch** (not main).
- Stderr empty.
- Bare origin has `refs/heads/feature-push` equal to the linked worktree HEAD.
- Local main HEAD may differ from the feature tip (feature is ahead).

## Side Effects

- Publishes the current checkout branch tip to origin (option R).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	branch := req.WtBranch
	if branch == "" {
		t.Fatal("WtBranch empty")
	}
	assert.Output(t, resp.Stdout, v2StdoutTemplate(pushConfirmLine(branch)))

	if req.OriginBare == "" {
		t.Fatal("OriginBare empty")
	}
	if req.WtDir == "" {
		t.Fatal("WtDir empty")
	}
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)

	// Must not only push main when cwd is the linked feature branch.
	featureSHA := revParseHEAD(t, req.WtDir)
	mainSHA := revParseHEAD(t, req.MainRepo)
	if featureSHA == mainSHA {
		t.Fatal("fixture expected feature tip != main HEAD")
	}
	originFeature := revParseRef(t, req.OriginBare, "refs/heads/"+branch)
	if originFeature != featureSHA {
		t.Fatalf("origin/%s should be feature tip %s, got %s", branch, featureSHA, originFeature)
	}
}
```
