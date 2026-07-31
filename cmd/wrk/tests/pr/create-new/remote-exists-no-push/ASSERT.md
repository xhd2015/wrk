## Expected Output

```
PR created
title set: Fix login
comment added
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stdout is new-PR shape **without** any `pushed …` line.
- Stderr empty.
- Origin `refs/heads/feature-pr` still equals the pre-run snapshot (local ahead not published).
- Fake `gh` called create + comment.

## Side Effects

- No ensure-push when remote head already exists.
- PR created + comment added.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
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

	assert.Output(t, resp.Stdout, v2StdoutTemplate(prCreatedStdout(prDefaultTitle, prDefaultURL)))
	if strings.Contains(resp.Stdout, "pushed ") {
		t.Fatalf("must not ensure-push when remote head exists; stdout=%q", resp.Stdout)
	}

	beforeBytes, err := os.ReadFile(filepath.Join(req.WorkRoot, "origin-feature-before"))
	if err != nil {
		t.Fatalf("read origin snapshot: %v", err)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/"+req.WtBranch)
	if after != before {
		t.Fatalf("origin/%s mutated under --pr when remote already existed: before %s after %s",
			req.WtBranch, before, after)
	}
	// Local is ahead of origin.
	local := revParseHEAD(t, req.WtDir)
	if local == after {
		t.Fatal("fixture expected local HEAD ahead of origin tip")
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "create")
	_ = assertGhSubcmdCalled(t, invocs, "comment")
}
```
