## Expected

- Exit code 0; stderr empty.
- Content (Branch/Commit/Status/Master/Remote) matches `wrk --status` from main.
- **Dir** lines use invocation cwd = external wt (`statusDirLine`): main is typically
  `../../myrepo` (≤2 ups), external block is `.` when appended.
- **Not** byte-equal to status-from-main when Dir differs (old equivalence dropped).

## Side Effects

- No nested shell; git status reporting only.
- `events.jsonl` may append (asserted under `events/`).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutMainStatusDirAware(t, req, resp, req.MainRepo, req.WtDir)

	ref := runStatusFromMain(t, req)
	if resp.Stdout == ref.Stdout {
		t.Fatalf("expected Dir-aware difference vs status-from-main when cwd is external wt; stdout:\n%s", resp.Stdout)
	}
}
```
