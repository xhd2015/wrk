## Expected Output

```text
.
tools/child
```

## Expected

- Exit code 0.
- Stdout is exactly `.\ntools/child\n`.
- Stderr is empty.

## Side Effects

- No repository files are changed.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	want := ".\ntools/child\n"
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant %q\ngot  %q", want, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
