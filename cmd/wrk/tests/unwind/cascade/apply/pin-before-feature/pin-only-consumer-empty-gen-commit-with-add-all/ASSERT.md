## Expected

- Exit code **0** (empty gen-commit after pinReady is not fatal with `--add-all`).
- Mid (agent-pro) freeHost:
  - Feature gen-commit subject `feat: add feature` on mid main history.
  - `FEATURE_WIP.md` landed on mid.
- Root consumer (after land):
  - `require example.com/agent-pro v0.0.2` (mid next after feature tag).
  - **no** external `replace` for agent-pro or dot-pkgs on Path go.mod.
  - go.mod/go.sum committed clean (no leftover pin porcelain).
- Cascade pin commits on root history: **no** external-style replace lines in
  pin commit trees (`./external/`, `../`, abs path) — pin isolation under
  `--add-all`.
- Linked root branch and `main` are **not diverged** (`rev-list --left-right
  --count main...branch` → `0 0`) — merge-back completed after empty gen-commit.
- Stdout/stderr must not leave the run as a hard failure solely for
  `no staged changes` when the worktree is clean after pinReady.

## Side Effects

- Mid peels/gen-commits/lands before (or interleaved with) root pin consumer peel.
- Root may emit pinReady/cascade pin auto-commits then skip AI gen-commit.
- `--merge-back` keeps linked worktrees; `--push` / `--sync` may run when clear.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		// Today RED: peel root hard-fails "no staged changes" under --add-all
		// after pinReady empties the index → no merge-back → diverged.
		t.Fatalf("P-empty pin-only-consumer: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.MainRepo == "" || req.DepPath == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo, DepPath (mid), and LeafModulePath required")
	}

	// Mid feature peel succeeded (production: agent-pro feature commit + land).
	assertMidFeatureLanded(t, req)

	// Root require mid @ next; drop external replaces for mid + leaf.
	assertConsumerRequireAndNoExternalReplace(t, req)
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			assertExternalReplaceDropped(t, filepath.Join(req.WtDir, "go.mod"), unwindDotPkgsModule)
		}
	}
	// Also check main after land when distinct.
	if req.MainRepo != "" {
		mainMod := filepath.Join(req.MainRepo, "go.mod")
		if _, err := os.Stat(mainMod); err == nil {
			// Soft if product pins only Path; prefer Path contract already checked.
			if goModHasReplace(t, mainMod, req.LeafModulePath) {
				t.Fatalf("main go.mod still has replace for %s after land:\n%s",
					req.LeafModulePath, readFile(t, mainMod))
			}
		}
	}

	hist := historyRepoForConsumer(t, req)
	assertGoModCommittedClean(t, hist)
	assertCascadePinCommitsNoExternalReplace(t, hist)
	// Prefer also scan linked Path history if still present and distinct.
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			assertCascadePinCommitsNoExternalReplace(t, req.WtDir)
			assertGoModCommittedClean(t, req.WtDir)
		}
	}

	assertLinkedBranchNotDivergedFromMain(t, req)

	// Sanity: at least one cascade pin appears on consumer history.
	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit on root history:\n%s", log)
	}
}
```
