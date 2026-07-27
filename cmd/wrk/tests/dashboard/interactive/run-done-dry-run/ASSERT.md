---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit **0**.
- `WRK_DASHBOARD_COMPOSE_ARGV_LOG` lists default DONE recipe:
  `--gen-commit-msg --add-all --commit --agent-runner=commandcode --done --sync --tag-next --push --reinstall-local --dry-run`
  (order free; space or newline separated tokens).
- Real compose dry-run evidence on stdout/stderr (`merge --ff-only` / `would:` / planned), not P2 snapshot-only.
- Linked worktree still present (dry-run zero remove apply).

## Side Effects

- Plan only; no worktree remove.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("run-done dry-run exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertComposeArgvRecipeDone(t, req, true /* dryRun */, true /* add-all: dirty */)
	assertDryRunComposeEvidence(t, resp, "done")
	assertLinkedWorktreeStillPresent(t, req)
}
```
