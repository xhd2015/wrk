
## Expected Output

```
would: go run ./script/foo/install
would: reinstall 1 binaries (0 skipped)
```

stderr:

```
notice: bin foo: preferring ./script/foo/install over ./cmd/foo
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above (plan lines uncolored).
- Stdout must **not** contain `go install ./cmd/foo` (script wins).
- Stderr is exactly the plain prefer-script notice (no ANSI under default pipe).
- Stub binary under GOBIN remains unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go run ./script/foo/install\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "go install ./cmd/foo")
	assertNoANSI(t, resp.Stdout)
	assertNoANSI(t, resp.Stderr)
	assertStderrExact(t, resp.Stderr, "notice: bin foo: preferring ./script/foo/install over ./cmd/foo\n")
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
