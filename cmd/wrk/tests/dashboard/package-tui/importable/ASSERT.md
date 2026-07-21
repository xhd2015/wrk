## Expected

- `go list github.com/xhd2015/wrk/wrkcli/tui` succeeds from the module root.
- Output is exactly the import path (trimmed), proving the package is part of the module.

## Errors

- Package missing / wrong path → `go list` non-zero → leaf fails (classic RED until implementer creates `wrkcli/tui`).

## Exit Code

- CLI `wrk -h` is incidental (expect 0); failure mode is package list, not help.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Harness run is only to satisfy shared Run; package surface is the contract.
	if err != nil {
		t.Fatalf("unexpected Run error for wrk -h: %v", err)
	}
	assertPackageListed(t, d, wrkcliTuiImportPath)
}
```
