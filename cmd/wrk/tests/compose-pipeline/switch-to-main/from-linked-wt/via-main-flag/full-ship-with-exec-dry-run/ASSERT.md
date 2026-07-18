## Expected

- Stderr must **not** contain `mutually exclusive`.
- Stderr must **not** claim `--exec` is invalid with `--main` / pipeline posts / `--reinstall-local`.
- Flag layer accepts `--main` + multi-stage + `--exec`.
- Exit 0 preferred with tag/reinstall plan evidence when dry-run succeeds.
- No done/merge-back plan; dry-run zero mutations when exit 0.

## Side Effects

- None required beyond dry-run zero mutations.

## Exit Code

- 0 preferred; non-zero only if not a flag-matrix reject of this combo.

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	assertNoMutexReject(t, se)
	if strings.Contains(se, "--exec is not valid") ||
		strings.Contains(se, "--exec is only valid") ||
		(strings.Contains(se, "--exec") && strings.Contains(se, "not valid with")) {
		t.Fatalf("--exec rejected with --main multi-stage compose; stderr=%q", se)
	}
	if resp.ExitCode != 0 {
		if strings.Contains(se, "mutually exclusive") ||
			strings.Contains(se, "unexpected") ||
			(strings.Contains(se, "--main") && strings.Contains(se, "not valid")) {
			t.Fatalf("exit %d unexpected for --main+posts+exec dry-run; stderr=%q", resp.ExitCode, se)
		}
	}
	if resp.ExitCode == 0 {
		if !strings.Contains(resp.Stdout, "1 tag planned") {
			t.Fatalf("expected tag-next plan under --main+exec; stdout=%q", resp.Stdout)
		}
		assertNotContains(t, resp.Stdout, "merge --ff-only")
		assertNotContains(t, resp.Stdout, "worktree remove")
		assertAPDryRunZeroMutationsLinked(t, req)
		assertStubPresentAP(t, filepath.Join(req.WorkRoot, "gobin"))
	}
}
```
