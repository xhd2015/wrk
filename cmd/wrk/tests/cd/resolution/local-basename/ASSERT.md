## Expected

- Exit 0; empty stdout (in-place).
- Follow-up targets **local** `workspace/myrepo`, not the saved projects path.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertFollowupCDLine(t, req, req.MainRepo)
	got := readFollowupFile(t, req.FollowupFile)
	if strings.Contains(got, resolvePath(t, req.SecondRepo)) {
		t.Fatalf("follow-up should use local dir, not saved project; got %q", got)
	}
}
```
