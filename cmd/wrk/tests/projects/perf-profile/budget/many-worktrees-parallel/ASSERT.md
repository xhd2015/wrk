---
label: slow, flaky
---
## Expected

After parallel worktree gathering (fixed worker pool + single main-repo goroutine):

- worktree_status_all duration under 100ms for 12 worktrees
- run_end total under 200ms for the whole --projects run

Serial baseline today (~190ms worktree aggregate, ~300ms run) must not pass these budgets.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	events := readProjectsPerfLog(t, req.ProjectsPerfLog)
	runMS := perfRunEndMS(events)
	wtMS, wtCount := perfPhaseTotalMS(events, "worktree_status_all")

	const maxWorktreeMS = int64(100)
	const maxRunMS = int64(200)

	if wtCount != 12 {
		t.Fatalf("worktree_status_all count: want 12, got %d", wtCount)
	}
	if wtMS >= maxWorktreeMS {
		t.Fatalf("worktree_status_all took %dms, want <%dms (serial per-worktree git status is too slow; need parallel pool)",
			wtMS, maxWorktreeMS)
	}
	if runMS >= maxRunMS {
		t.Fatalf("run_end took %dms, want <%dms", runMS, maxRunMS)
	}
	t.Logf("perf budget OK: run_end=%dms worktree_status_all=%dms", runMS, wtMS)
}
```