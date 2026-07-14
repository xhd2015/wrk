## Expected Output

```text
Multiple projects match "myrepo":
  1) <sorted-path-1>
  2) <sorted-path-2>
Select [1-2]:
```

## Expected

- Exit code 0.
- Stdout contains one status block for the **selected** saved repo (`zzz/myrepo`, index 2).
- Stderr shows the ambiguous-basename prompt (or simulated via `WRK_BASENAME_CONFIRM=1`).

## Side Effects

- TTY prompt shown; stdin index selects candidate; status run against selected path only.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	sorted := sortedSavedPaths(t, req.MainRepo, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "myrepo":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
Select [1-2]:
</contains>`
	assert.Output(t, resp.Stderr, tmpl)

	if got := statusOutputBlockCount(resp.Stdout); got != 1 {
		t.Fatalf("expected 1 status block, got %d:\n%s", got, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, statusBlockTemplate(t, req.SelectedSavedRepo, ".", "clean"))
}
```