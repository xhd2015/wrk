## Expected Output

```text
Dir:          .
…
Dir:          external/aaa-dep
…
Dir:          external/zzz-dep
…
```

## Expected

- Exit code 0; stderr empty.
- **Exactly 3** `Dir:` blocks (consumer + two external deps).
- Dir lines include `.`, `external/aaa-dep`, and `external/zzz-dep` (via statusDirLine).
- No `---- external ----` header.
- Master on all three linked blocks (consumer main / depA main / depZ main).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := statusOutputBlockCount(resp.Stdout); got != 3 {
		t.Fatalf("expected 3 status blocks (consumer + 2 external deps), got %d:\n%s", got, resp.Stdout)
	}
	assertNoExternalSectionHeader(t, resp.Stdout)

	dirA := statusDirLine(t, req.RepoDir, req.ExternalWtDir)
	dirZ := statusDirLine(t, req.RepoDir, req.ExternalWtDir2)
	assertStdoutHasDirLine(t, resp.Stdout, ".")
	assertStdoutHasDirLine(t, resp.Stdout, dirA)
	assertStdoutHasDirLine(t, resp.Stdout, dirZ)

	// Match SETUP branch literals (do not use TaskDesc — it becomes CLI --task).
	branchA := req.Wt2Branch
	if branchA == "" {
		branchA = "dep-a-" + wrkDate
	}
	branchZ := "dep-z-" + wrkDate

	// Path-sorted: consumer, then external/aaa-dep, then external/zzz-dep.
	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		linkedScanBlock(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		linkedScanBlock(t, req.RepoDir, req.DepPath, req.ExternalWtDir, branchA, "clean"),
		linkedScanBlock(t, req.RepoDir, req.DepsDepPath, req.ExternalWtDir2, branchZ, "clean"),
	))
}
```
