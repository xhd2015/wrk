## Expected Output

```text
.
```

## Expected

- Exit code 0.
- Stdout is exactly `.\n`.
- Stderr is empty.

## Side Effects

- No repository files are changed.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stdout != ".\n" {
		t.Fatalf("stdout mismatch:\nwant %q\ngot  %q", ".\n", resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
