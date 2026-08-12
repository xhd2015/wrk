## Expected

- Exit code 0.
- Free monorepo after apply:
  - Free main has next content in `pkg.go` (auto-commit landed uncommitted WIP).
  - Local free root tag `v0.0.2` exists (deferred or cascade tag-next).
  - Bare origin has free tag `v0.0.2` when push completes (soft if push defers).
- Consumer after apply (`req.MainRepo`):
  - `require example.com/dot-pkgs v0.0.2` (**must** re-pin after free next tag)
  - **must not** require `example.com/dot-pkgs/cmd`
  - **no** droppable external replace for free root
  - cascade pin commit for free root @ next present when require bumped
- Combined output must **not** contain `unknown revision`.

## Side Effects

- Free peel may be deferred (pinConsumer without freeHost when NextTag empty at
  plan); auto-commit lands uncommitted free feature; deferred tags create free
  next; **desired product** re-pins consumers to that next version.
- `--done` may remove free external worktree (OK; assert on free main).

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		t.Fatalf("CS-repin deferred free tag must re-pin consumer: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("CS-repin: unknown revision after apply\ncombined:\n%s", out)
	}

	// Free feature must have landed on free main (auto-commit + merge-back).
	pkg := readFile(t, filepath.Join(req.SecondRepo, "pkg.go"))
	if !strings.Contains(pkg, unwindApplyNextTag) {
		t.Fatalf("CS-repin: free main pkg.go should include next-tag content after peel auto-commit; got:\n%s", pkg)
	}

	// Free next tag must exist (deferred tags or cascade).
	if !tagRefExists(t, req.SecondRepo, req.ExpectedPinVersion) {
		t.Fatalf("CS-repin: free main missing root tag %s (deferred tag expected)\nlocal tags:\n%s",
			req.ExpectedPinVersion, gitOutputIsolated(t, req.SecondRepo, "tag", "-l"))
	}
	if req.OriginBare != "" && !remoteTagExists(t, req.OriginBare, req.ExpectedPinVersion) {
		t.Logf("CS-repin note: origin missing %s after --push (local tag present)",
			req.ExpectedPinVersion)
	}

	// Core desired product: consumer re-pinned to free next after deferred tag.
	// Crime-scene bug left require at Latest (v0.0.1) while free was tagged next.
	assertConsumerRequireAndNoExternalReplace(t, req)
	assertConsumerPinnedRootOnly(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinCommitPresent(t, hist, req.LeafModulePath, req.ExpectedPinVersion)
	assertGoModCommittedClean(t, hist)
}
```
