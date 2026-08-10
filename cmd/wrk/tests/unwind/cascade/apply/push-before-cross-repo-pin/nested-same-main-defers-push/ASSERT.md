## Expected

- Exit code 0.
- Free monorepo:
  - Local root tag `v0.0.2` on free main.
  - Bare origin has free `main` + root tag `v0.0.2` when push completes.
- **Order (core C-PUSH1):** free root `tag-next` → free `pushed main →` →
  cross-repo `pin app <- dot-pkgs @ v0.0.2` (push must not wait for nested
  same-main pin/tag on free).
- Consumer (`req.MainRepo` / example.com/app):
  - `require example.com/dot-pkgs v0.0.2`
  - no droppable external replace for free
  - cascade pin commit for free root @ next
- Combined output must **not** contain `unknown revision`.

## Side Effects

- Peel lands free multi-module WT; cascade tags free root, must **publish** free
  before pin of app (network pin after droppable external replace).
- Nested free/cmd may pin (keep-local) and tag after free root publish.
- `--done` may remove free external worktree (OK).
- Offline file:// modproxy supplies free root old+next (order assert is the
  RED surface; production without push fails tidy with unknown revision).

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
		t.Fatalf("C-PUSH1: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, LeafModulePath required")
	}
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("C-PUSH1: unknown revision after apply\ncombined:\n%s", out)
	}

	// Core ordering: free tag → free push → cross-repo pin of free @ next.
	assertFreePushBeforeCrossRepoPinOfFree(t, out)

	// Free root next tag local + origin.
	if !tagRefExists(t, req.SecondRepo, req.ExpectedPinVersion) {
		t.Fatalf("C-PUSH1: free main missing root tag %s", req.ExpectedPinVersion)
	}
	if req.OriginBare != "" {
		assertOriginMainEqualsLocalMain(t, req.SecondRepo, req.OriginBare)
		if !remoteTagExists(t, req.OriginBare, req.ExpectedPinVersion) {
			t.Fatalf("C-PUSH1: %s should exist on bare origin after free push",
				req.ExpectedPinVersion)
		}
	}

	// Consumer require bumped; replace dropped; pin commit present.
	assertRequireBumped(t, req)
	goMod := filepath.Join(req.MainRepo, "go.mod")
	if goModHasReplace(t, goMod, req.LeafModulePath) {
		t.Fatalf("C-PUSH1: consumer must drop external replace for %s\ngo.mod:\n%s",
			req.LeafModulePath, readFile(t, goMod))
	}
	assertCascadePinCommitPresent(t, req.MainRepo, req.LeafModulePath, req.ExpectedPinVersion)
	assertGoModCommittedClean(t, req.MainRepo)
}
```
