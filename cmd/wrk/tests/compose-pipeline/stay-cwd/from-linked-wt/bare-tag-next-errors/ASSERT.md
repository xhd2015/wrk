## Expected

- Non-zero exit.
- Stderr names `--tag-next`.
- Stderr indicates main repository / main checkout requirement (not a silent main tag apply).
- Must not create v0.0.2 on main.
- Must not print successful apply-only tag lines without the gate error.

## Errors

- `--tag-next` without done/merge-back requires activeRoot main (cwd main).

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
		t.Fatalf("expected non-zero for --tag-next from linked wt; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--tag-next") && !strings.Contains(se, "tag-next") {
		t.Fatalf("stderr should name --tag-next, got %q", se)
	}
	lower := strings.ToLower(se)
	if !strings.Contains(lower, "main") {
		t.Fatalf("stderr should mention main repository requirement, got %q", se)
	}
	if tagRefExistsAP(t, req.MainRepo, "v0.0.2") {
		t.Fatal("must not tag main when --tag-next is invoked from linked worktree")
	}
}
```
