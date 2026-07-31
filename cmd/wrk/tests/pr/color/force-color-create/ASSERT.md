## Expected Output

```
PR created   (green)
title set: Fix login   (green token only)
comment added   (green)
https://github.com/acme/app/pull/42   (plain)
```

## Expected

- Exit code 0.
- Stdout success tokens `PR created`, `title set`, and `comment added` are green ANSI (`\x1b[32m`…`\x1b[0m`).
- Title value and URL stay plain (uncolored).
- Stderr empty.
- Fake `gh` create + comment called.

## Side Effects

- Same as create-new with remote present (no push of tip); color only.

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
	assertEmptyStderr(t, resp.Stderr)

	// v3: green success tokens; plain title value + URL.
	tmpl := `---
version: 3
---
<ansi-color green>PR created</ansi-color>
<ansi-color green>title set</ansi-color>: Fix login
<ansi-color green>comment added</ansi-color>
https://github.com/acme/app/pull/42
`
	assert.Output(t, resp.Stdout, tmpl)

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "create")
	_ = assertGhSubcmdCalled(t, invocs, "comment")
}
```
