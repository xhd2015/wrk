## Expected

- Exit code 0 (does not error with confirmation-required; does not hang).
- No `Proceed?` on stdout/stderr.
- Stderr/stdout must **not** contain `stdin is not a terminal` / `cannot prompt`.
- Primary dry-run plan present (`merge --ff-only` + remove commands).
- Zero mutations (same as alone).

## Side Effects

- None.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("dry-run without -y must exit 0 (no confirm required); exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)
	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)
	assertDoneDryRunZeroMutations(t, req)
}
```
