## Expected

- Exit code 0.
- Stdout includes `would: git push origin main` and `would: git push --force-with-lease origin <WtBranch>`.
- No `pushed` confirm lines.
- Origin refs and worktree unchanged.

## Exit Code

- 0

```go
import (
	"fmt"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	if !strings.Contains(resp.Stdout, "would: git push origin main") {
		t.Fatalf("missing would: main push\n%s", resp.Stdout)
	}
	wantLease := fmt.Sprintf("would: git push --force-with-lease origin %s", req.WtBranch)
	if !strings.Contains(resp.Stdout, wantLease) {
		t.Fatalf("missing %q\n%s", wantLease, resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "pushed main →") {
		t.Fatalf("dry-run must not print pushed confirm; stdout=%q", resp.Stdout)
	}

	wtSHA := revParseHEAD(t, req.WtDir)
	assertOriginBranchEquals(t, req.OriginBare, req.WtBranch, wtSHA)
	mainSHA := revParseHEAD(t, req.MainRepo)
	assertOriginBranchEquals(t, req.OriginBare, "main", mainSHA)
	if mainSHA == wtSHA {
		t.Fatal("dry-run must not land the worktree into main")
	}
}
```
