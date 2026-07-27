## Expected Output

```text
.
```

## Expected

- Exit code 0.
- Stdout is exactly `.\n`.
- Stderr is empty.

## Side Effects

- Repo paths reported for the saved project resolved via basename fallback, not cwd.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != ".\n" {
		t.Fatalf("stdout mismatch:\nwant %q\ngot  %q", ".\n", resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```