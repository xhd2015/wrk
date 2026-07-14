## Expected

- Exit code 0; stderr empty.
- Stdout equals running `wrk --status` from the main repo (equivalence promise).
- Implies multi-block main view (`.` + `Remote:`, appended external block with abs `Dir` and `Master:`).

## Side Effects

- No nested shell; git status reporting only.
- `events.jsonl` may append (asserted under `events/`).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutEqualsMainStatus(t, req, resp)
}
```