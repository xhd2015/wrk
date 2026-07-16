## Expected Output

```
would: reinstall 0 binaries (0 skipped)
```

stderr:

```
warning: bin foo: ambiguous under cmd (./cmd/foo, ./cmd/nested/foo); skipping
```

## Expected

- Exit code 0 (diagnostics non-fatal).
- Stdout is only the summary with N=0 M=0 (no install/skip rows for omitted bin).
- Stdout must not contain `go install` for `foo`.
- Stderr plain warning with sorted paths (no ANSI under pipe).
- Stub binary under GOBIN remains unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: reinstall 0 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "go install")
	assertNotContains(t, resp.Stdout, "go run")
	assertNoANSI(t, resp.Stdout)
	assertNoANSI(t, resp.Stderr)
	assertStderrExact(t, resp.Stderr,
		"warning: bin foo: ambiguous under cmd (./cmd/foo, ./cmd/nested/foo); skipping\n")
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
