## Expected Output

```
v0.0.1
```

## Expected

- Exit code 0.
- Stdout is exactly `v0.0.1` plus trailing newline.
- Stderr is empty.

## Side Effects

- Read-only; no `events.jsonl` under `WRK_HOME`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate("v0.0.1\n"))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertFileNotExists(t, eventsJSONLPath(req.WrkHome))
}
```