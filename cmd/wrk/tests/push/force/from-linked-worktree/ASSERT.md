## Expected Output

```
pushed feature-push → origin/feature-push
```

## Expected

- Exit code 0.
- Stdout: confirm line for the **worktree branch** (not main); still `pushed …` wording.
- Stderr empty.
- Bare origin has `refs/heads/feature-push` equal to the linked worktree HEAD.
- Feature tip differs from main HEAD (fixture integrity).

## Side Effects

- Publishes current checkout branch tip to origin with force-with-lease (option R + force).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
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
	if strings.Contains(resp.Stdout, "force-pushed") {
		t.Fatalf("confirm line must stay 'pushed …'; got %q", resp.Stdout)
	}

	if req.OriginBare == "" {
		t.Fatal("OriginBare empty")
	}
	if req.WtDir == "" {
		t.Fatal("WtDir empty")
	}
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)

	featureSHA := revParseHEAD(t, req.WtDir)
	mainSHA := revParseHEAD(t, req.MainRepo)
	if featureSHA == mainSHA {
		t.Fatal("fixture expected feature tip != main HEAD")
	}
}
```
