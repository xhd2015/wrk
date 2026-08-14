## Expected

- Exit code 0.
- Last non-empty stdout line is the **project** worktree path (`pwd` result), not the external path.
- Create path and external path also appear as stdout lines.
- `cmd.Dir` of `--exec` is the new project WT.

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wt := createBringDefaultWT(req)
	ext1 := createBringExternalPath(wt, "mydep1")
	assertFileExists(t, wt)
	assertFileExists(t, ext1)
	if !createBringStdoutHasLine(resp.Stdout, wt) {
		t.Fatalf("stdout should include create path %q; got %q", wt, resp.Stdout)
	}
	if !createBringStdoutHasLine(resp.Stdout, ext1) {
		t.Fatalf("stdout should include external %q; got %q", ext1, resp.Stdout)
	}

	s := strings.TrimSuffix(resp.Stdout, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || lines[len(lines)-1] == "" {
		t.Fatalf("stdout empty; want last line pwd == %q\n%s", wt, resp.Stdout)
	}
	got := lines[len(lines)-1]
	if got != wt {
		t.Fatalf("exec pwd cmd.Dir: want project WT %q, got %q (must not be external %q)\nstdout:\n%s", wt, got, ext1, resp.Stdout)
	}
}
```
