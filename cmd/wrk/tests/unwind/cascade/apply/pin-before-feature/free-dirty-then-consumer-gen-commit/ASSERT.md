## Expected

- Exit code 0.
- Free-first ship:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- Consumer:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** external replace for that module
  - cascade pin commit present; **before** feature gen-commit (`feat: add feature`)
  - `FEATURE_WIP.md` landed
- go.mod/go.sum committed clean on consumer history checkout.

## Side Effects

- Free peel may gen-commit leaf dirt then merge-back/land before tag.
- Pin auto-commit selective go.mod/go.sum; feature gen-commit separate (D7).
- `--merge-back` keeps worktrees; `--push` publishes free leaf when clear.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		t.Fatalf("T2 free-dirty-then-consumer-gen-commit: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" {
		t.Fatal("SecondRepo and MainRepo required")
	}

	// Free-first: leaf tagged at next before/with consumer pin completion.
	assertLeafMainAdvancedAndTagged(t, req)

	// Consumer require bump + drop external replace.
	assertConsumerRequireAndNoExternalReplace(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-20")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("expected feature subject %q in consumer log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```
