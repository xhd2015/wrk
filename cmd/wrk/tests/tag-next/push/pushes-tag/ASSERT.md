## Expected

- Exit code 0.
- Local tag `v0.0.2` exists at HEAD.
- Bare `origin` has `refs/tags/v0.0.2`.
- Stdout footer `1 tag created`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertContains(t, resp.Stdout, "1 tag created")

	if !tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 should exist locally after --push apply")
	}
	if !remoteTagExists(t, req.OriginBare, "v0.0.2") {
		t.Fatal("v0.0.2 should exist on bare origin after --push")
	}
}
```