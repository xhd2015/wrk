## Expected

- Exit 0 (or non-mutex failure only if dry-run cannot plan exec — still must accept flags).
- Stderr must **not** contain `mutually exclusive`.
- Stderr must **not** claim `--exec` is invalid with `--done` / pipeline posts / `--reinstall-local`.
- Done merge plan present; reinstall plan present when applicable.
- `--exec` is the last requested stage (after reinstall in the pipeline model).
- Dry-run: no real `true` side effects required; zero git mutations.

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
		t.Fatalf("--exec rejected with done multi-stage compose; stderr=%q", se)
	}
	if resp.ExitCode != 0 {
		// Flag accept is the RED unlock; allow only non-mutex later-stage issues.
		if strings.Contains(se, "mutually exclusive") || strings.Contains(se, "unexpected") {
			t.Fatalf("exit %d unexpected for done+posts+exec dry-run; stderr=%q", resp.ExitCode, se)
		}
	}
	if resp.ExitCode == 0 {
		assertContains(t, resp.Stdout, "merge --ff-only "+req.WtBranch)
		assertAPDryRunZeroMutationsLinked(t, req)
		assertStubPresentAP(t, filepath.Join(req.WorkRoot, "gobin"))
	}
}
```
