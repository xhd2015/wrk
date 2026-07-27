## Expected

- Exit code 0.
- Stdout includes MergeBack DryRun planned commands for ahead + remove
  (`merge --ff-only <WtBranch>`, `worktree remove`, `branch -D <WtBranch>`).
- No confirm prompt (`Proceed?`) and no non-TTY confirm errors.
- Stdout does **not** include post-stage vocabulary:
  - no `would: synced:` / `synced:`
  - no `tag planned` / `tag created` / `tagged `
  - no `would: git push` / `pushed main`
- Side effects: **zero mutations** — wt still linked; main HEAD unchanged;
  `feature-work` not on main; no `v0.0.2` tag.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)
	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)

	// Alone: no post stages.
	assertNotContains(t, resp.Stdout, "would: synced:")
	assertNotContains(t, resp.Stdout, "\nsynced:")
	if strings.Contains(resp.Stdout, "synced:") && !strings.Contains(resp.Stdout, "would:") {
		// bare synced: without would is real apply — forbidden
		t.Fatalf("alone dry-run must not run real sync; stdout=%q", resp.Stdout)
	}
	assertNotContains(t, resp.Stdout, "tag planned")
	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "tagged ")
	assertNotContains(t, resp.Stdout, "would: git push")
	assertNotContains(t, resp.Stdout, "pushed main")

	assertDoneDryRunZeroMutations(t, req)
}
```
