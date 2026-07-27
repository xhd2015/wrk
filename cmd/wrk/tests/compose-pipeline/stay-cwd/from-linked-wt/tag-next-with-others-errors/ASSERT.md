## Expected

- Non-zero exit.
- Stderr names `--tag-next` and main requirement.
- No successful silent tag-on-main from linked worktree.
- No v0.0.2 created.

## Errors

- Multi-stage with `--tag-next` still requires main activeRoot when done/merge-back is absent.

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
		t.Fatalf("expected non-zero for --tag-next multi-stage from linked wt; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--tag-next") && !strings.Contains(se, "tag-next") {
		t.Fatalf("stderr should name --tag-next, got %q", se)
	}
	if !strings.Contains(strings.ToLower(se), "main") {
		t.Fatalf("stderr should mention main, got %q", se)
	}
	if tagRefExistsAP(t, req.MainRepo, "v0.0.2") {
		t.Fatal("must not create v0.0.2 from linked-wt multi-stage without done")
	}
}
```
