## Expected Output

```
would: go install ./cmd/foo
would: reinstall 1 binaries (0 skipped)
```

stderr:

```
warning: bin foo: ambiguous under script (./script/foo/install, ./script/x/foo/install); skipping
```

## Expected

- Exit code 0.
- Stdout plans `go install ./cmd/foo` only (not go run script).
- Stderr plain ambiguous-script warning with sorted paths.
- No ANSI under default pipe.
- Stub binary under GOBIN remains unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/foo\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "go run")
	assertNoANSI(t, resp.Stdout)
	assertNoANSI(t, resp.Stderr)
	assertStderrExact(t, resp.Stderr,
		"warning: bin foo: ambiguous under script (./script/foo/install, ./script/x/foo/install); skipping\n")
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
