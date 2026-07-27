## Expected

- One `list_linked` phase per project after deduplicating `ListLinked` (not separate `list_linked_skip` + `list_linked_summary`).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	events := readProjectsPerfLog(t, req.ProjectsPerfLog)
	got := perfListLinkedPhaseCount(events)
	if got != 1 {
		t.Fatalf("list_linked phase count: want 1 (shared ListLinked), got %d — duplicate list_linked_skip + list_linked_summary still present", got)
	}
}
```