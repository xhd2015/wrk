## Expected Output

```
# stderr:
warning: title ignored (PR already exists); existing title: Fix login
# stdout:
comment added
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stderr contains the title-ignored warning with existing title `Fix login` (and `warning:` prefix).
- Stdout is comment-added + URL only (no `PR created`, no `title set`, no `pushed`).
- Fake `gh`: `pr list` and `pr comment` called; **`pr create` not called**.

## Side Effects

- Additive comment only; existing PR title unchanged (no create/edit title).

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

	assert.Output(t, resp.Stdout, v2StdoutTemplate(prExistingStdout(prDefaultURL)))
	if strings.Contains(resp.Stdout, "PR created") || strings.Contains(resp.Stdout, "title set") {
		t.Fatalf("existing PR path must not print create/title-set tokens; stdout=%q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "pushed ") {
		t.Fatalf("must not push when remote head exists; stdout=%q", resp.Stdout)
	}

	se := resp.Stderr
	if !strings.Contains(se, "warning:") {
		t.Fatalf("stderr should include warning: prefix, got %q", se)
	}
	if !strings.Contains(se, "title ignored") {
		t.Fatalf("stderr should say title ignored, got %q", se)
	}
	if !strings.Contains(se, prExistingTitle) {
		t.Fatalf("stderr should include existing title %q, got %q", prExistingTitle, se)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	_ = assertGhSubcmdCalled(t, invocs, "comment")
	assertGhSubcmdNotCalled(t, invocs, "create")
}
```
