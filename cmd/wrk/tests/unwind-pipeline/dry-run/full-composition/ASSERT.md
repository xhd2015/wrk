## Expected

- Exit 0 with a newline-terminated phase dry-run plan.
- Peel/lane display uses relative path `external/dep` (not bare `dep` alone).
- Plan includes gen-commit, merge-back, reinstall-local, push, sync for the dep lane.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertExit0(t, resp)
	out := resp.Stdout
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("stdout lacks final newline: %q", out)
	}
	if !strings.Contains(out, "external/dep") {
		t.Fatalf("want lane/peel display external/dep\nstdout:\n%s", out)
	}
	assertContainsInOrder(t, out,
		"external/dep",
		"would: gen-commit-msg",
		"would: commit",
		"would: merge-back",
		"would: push",
		"would: sync",
	)
	if !strings.Contains(out, "would: reinstall-local") &&
		!strings.Contains(out, "would: reinstall local binaries") {
		t.Fatalf("want reinstall in dry-run plan\nstdout:\n%s", out)
	}
	if got := git(t, req.DepMain, "rev-parse", "HEAD"); got != req.BeforeDep {
		t.Fatalf("dry-run mutated dep: %s != %s", got, req.BeforeDep)
	}
	if got := git(t, req.MainRepo, "rev-parse", "HEAD"); got != req.BeforeMain {
		t.Fatalf("dry-run mutated main: %s != %s", got, req.BeforeMain)
	}
}
```
