
## Expected Output

```
go install ./foo
reinstalled 1, skipped 0, failed 0
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above (trailing newline on last line).
- Progress uses **post-re-root** path `go install ./foo` (not
  `go install ./cmd/foo`).
- Summary: `reinstalled 1, skipped 0, failed 0` (must not be `failed 1` from
  “main module does not contain package …/cmd/foo”).
- No `would:` dry-run vocabulary and no `# module` / `across` dry-run headers.
- `$GOBIN/foo` is a real executable (not the stub) and prints `foo-ok` — proves
  `go install` ran with `Dir` = nested `cmd/` module root.

## Side Effects

- Real `go install ./foo` with `Dir=<nearest go.mod under plan ModuleRoot>` and
  `GOBIN=BinDir` replaces the stub.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "" +
		"go install ./foo\n" +
		"reinstalled 1, skipped 0, failed 0\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "would:")
	assertNotContains(t, resp.Stdout, "# module")
	assertNotContains(t, resp.Stdout, "across")
	// Explicit anti-regression: unre-rooted parent RelPath must not be used.
	assertNotContains(t, resp.Stdout, "go install ./cmd/foo")
	assertBinNotStub(t, req.BinDir, "foo")
	assertBinExecutable(t, req.BinDir, "foo")
	assertBinRuns(t, req.BinDir, "foo", "foo-ok")
}
```
