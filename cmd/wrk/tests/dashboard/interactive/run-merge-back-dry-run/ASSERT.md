## Expected

- Exit **0**.
- Compose argv log: `--gen-commit-msg --commit --agent-runner=commandcode --merge-back --sync --tag-next --push --reinstall-local --dry-run` (no `--done`).
- Dry-run compose evidence; **no** `worktree remove` in plan output.
- Linked worktree still on disk.

## Side Effects

- Plan only; worktree kept.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("run-merge-back dry-run exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertComposeArgvRecipeMergeBack(t, req, true /* dryRun */)
	assertDryRunComposeEvidence(t, resp, "merge-back")
	assertLinkedWorktreeStillPresent(t, req)
}
```
