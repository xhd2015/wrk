## Expected

- Exit code **0**.
- Free multi-module cascade still ships (precondition for the sync hole):
  - Free main has cascade pin commit for free root / monorepo pin path
    (`wrk: cascade pin` mentioning free module).
  - Free root next tag `v0.0.2` exists on free main.
- Free linked worktree **kept** (`--merge-back`): `req.DepsLinkedWtDir` exists.
- **Post-cascade sync contract (desired):** free linked checkout tracks free main:
  - `rev-parse HEAD` of free linked WT == free main (`req.SecondRepo`) HEAD.
  - `rev-list --left-right --count main...<free-branch>` → `0 0`
    (no commits only on main, none only on branch).
- Combined output must not leave the free WT needing fast-forward solely because
  cascade pin landed after peel-time sync.

## Side Effects

- Peel free (gen-commit + merge-back + peel-time sync) then cascade tag/pin/push.
- Free external WT remains (unlike C1 `--done` which may remove it).
- Consumer pin / replace-drop may run; not the primary assert surface here.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"os"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		t.Fatalf("C1-sync free multi-module merge-back+sync: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.DepsLinkedWtDir == "" || req.WtBranch == "" {
		t.Fatal("SecondRepo (free main), DepsLinkedWtDir (free linked WT), WtBranch required")
	}
	if req.LeafModulePath == "" {
		t.Fatal("LeafModulePath (free root module) required")
	}

	// Free linked WT must still exist after --merge-back.
	if st, err := os.Stat(req.DepsLinkedWtDir); err != nil || !st.IsDir() {
		t.Fatalf("free linked WT must remain after --merge-back: %s: %v", req.DepsLinkedWtDir, err)
	}

	// Precondition: cascade advanced free main with a pin commit (the +1 hole).
	// Without a free-main pin, HEAD equality can pass vacuously after land alone.
	assertCascadePinCommitPresent(t, req.SecondRepo, req.LeafModulePath, req.ExpectedPinVersion)
	if !tagRefExists(t, req.SecondRepo, req.ExpectedPinVersion) {
		// Nested path-scoped root tag is v0.0.2 on free multi-module fixtures.
		if !tagRefExists(t, req.SecondRepo, unwindApplyNextTag) {
			t.Fatalf("free main missing root next tag %s (or %s)\nlocal tags:\n%s",
				req.ExpectedPinVersion, unwindApplyNextTag,
				gitOutputIsolated(t, req.SecondRepo, "tag", "-l"))
		}
	}

	freeMain := req.SecondRepo
	freeWT := req.DepsLinkedWtDir
	mainHEAD := revParseHEAD(t, freeMain)
	wtHEAD := revParseHEAD(t, freeWT)
	short := func(sha string) string {
		sha = strings.TrimSpace(sha)
		if len(sha) >= 7 {
			return sha[:7]
		}
		return sha
	}
	subj := func(repo, sha string) string {
		return strings.TrimSpace(gitOutputIsolated(t, repo, "log", "-1", "--format=%s", sha))
	}
	lr := strings.TrimSpace(gitOutputIsolated(t, freeMain, "rev-list", "--left-right", "--count",
		"main..."+req.WtBranch))

	// Desired: peel-time sync is not enough; post-cascade free linked WT tracks main.
	if mainHEAD != wtHEAD {
		t.Fatalf("free linked WT must track free main after --unwind --sync (post-cascade pin):\n"+
			"  free main (%s) HEAD=%s %s\n"+
			"  free WT   (%s) HEAD=%s %s\n"+
			"  left-right main...%s = %s\n"+
			"  free main log:\n%s\n  free branch log:\n%s\ncombined:\n%s",
			freeMain, short(mainHEAD), subj(freeMain, mainHEAD),
			freeWT, short(wtHEAD), subj(freeWT, wtHEAD),
			req.WtBranch, lr,
			gitOutputIsolated(t, freeMain, "log", "--oneline", "--decorate", "-12", "main"),
			gitOutputIsolated(t, freeMain, "log", "--oneline", "--decorate", "-12", req.WtBranch),
			out,
		)
	}

	// Also lock branch vs main: zero unique commits either side (status identical).
	parts := strings.Fields(lr)
	if len(parts) != 2 || parts[0] != "0" || parts[1] != "0" {
		t.Fatalf("free branch must not lag main after --sync (want left-right 0 0, got %q)\n"+
			"main HEAD=%s wt HEAD=%s\nmain log:\n%s\nbranch log:\n%s",
			lr, short(mainHEAD), short(wtHEAD),
			gitOutputIsolated(t, freeMain, "log", "--oneline", "--decorate", "-12", "main"),
			gitOutputIsolated(t, freeMain, "log", "--oneline", "--decorate", "-12", req.WtBranch))
	}
}
```
