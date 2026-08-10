## Expected

- Exit code 0.
- Free monorepo free-first ship:
  - Free main advanced with root next content in `pkg.go`.
  - Local root tag `v0.0.2` exists on free main (may not be HEAD if free/cmd pin
    commit landed after root tag).
  - Local nested tag `cmd/v0.0.2` exists on free main after free/cmd tag-next.
  - Bare origin has free `main` + root tag `v0.0.2` when push completes.
- **Order** (when apply logs cascade steps): free root `tag-next` before pin of
  free/cmd ← free, and free root tag-next before consumer pin of free @ next.
- Consumer after apply (`req.MainRepo`):
  - `require example.com/dot-pkgs v0.0.2`
  - **must not** require `example.com/dot-pkgs/cmd` at any version
  - **no** droppable external replace for free root
  - pin log: exactly **one** `pin root <- dot-pkgs` line (required pairing only;
    cartesian root+cmd spam is the multi-module bug surface)
  - cascade pin commit present for free root @ next when replace was dropped
- Combined output must **not** contain `unknown revision`.

## Side Effects

- Peel lands free multi-module WT; cascade tags free root then pins free/cmd
  (keep-local `=> ../`) then tags free/cmd; pins consumer free root only.
- `--done` may remove free external worktree (OK).
- Offline file:// modproxy supplies free root old+next only (no free/cmd@next).

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
		t.Fatalf("C1 free multi-module cmd + consumer root-only: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" || req.NestedModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, LeafModulePath, NestedModulePath required")
	}
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("C1: unknown revision after apply\ncombined:\n%s", out)
	}

	// Free root content advanced after land.
	pkg := readFile(t, filepath.Join(req.SecondRepo, "pkg.go"))
	if !strings.Contains(pkg, unwindApplyNextTag) {
		t.Fatalf("C1: free main pkg.go should include next-tag content; got:\n%s", pkg)
	}
	// Root next tag exists (not necessarily HEAD — free/cmd pin may commit after).
	if !tagRefExists(t, req.SecondRepo, req.ExpectedPinVersion) {
		t.Fatalf("C1: free main missing root tag %s", req.ExpectedPinVersion)
	}
	// Nested free/cmd next tag after free-first pin of free then tag-next cmd.
	if !tagRefExists(t, req.SecondRepo, freeMultiCmdNextTag) {
		// Tolerate tagscope spelling that still carries cmd/ + v0.0.2.
		if !tagRefExists(t, req.SecondRepo, "cmd/"+unwindApplyNextTag) {
			t.Fatalf("C1: free main missing nested cmd next tag (want %s)\nlocal tags:\n%s",
				freeMultiCmdNextTag, gitOutputIsolated(t, req.SecondRepo, "tag", "-l"))
		}
	}
	if req.OriginBare != "" {
		assertOriginMainEqualsLocalMain(t, req.SecondRepo, req.OriginBare)
		if !remoteTagExists(t, req.OriginBare, req.ExpectedPinVersion) {
			t.Fatalf("C1: %s should exist on bare origin after --tag-next --push",
				req.ExpectedPinVersion)
		}
	}

	// Free-first order when cascade logs are present.
	if hasCascadeTagNext(out, req.LeafModulePath) || strings.Contains(out, "tag-next "+req.LeafModulePath) {
		// Root tag before free/cmd pin of free (intra free monorepo free-first).
		if strings.Contains(out, "pin ") && strings.Contains(out, freeMultiCmdModule) {
			assertContainsInOrder(t, out,
				"tag-next "+req.LeafModulePath,
				freeMultiCmdModule,
			)
		}
		// Root tag before consumer pin of free @ next.
		assertFreeTagNextBeforeConsumerPinOfFree(t, out)
	}

	// Pin log: only consumer ← free root pairing (not cartesian root+cmd spam).
	assertExactlyOnePinLine(t, out, labelRoot, labelDotPkgs)

	// Consumer go.mod: root bumped; nested free/cmd must stay absent; replace dropped.
	assertConsumerPinnedRootOnly(t, req)

	// Cascade pin auto-commit for free root when replace was dropped.
	assertCascadePinCommitPresent(t, req.MainRepo, req.LeafModulePath, req.ExpectedPinVersion)
	assertGoModCommittedClean(t, req.MainRepo)
}
```