## Expected

- Exit code **0**.
- Stdout is exactly the resolved absolute path of `repo-a` (single line + trailing `\n`).
- `projects.json` contains exactly one entry for that main with `source: "scan"`.

## Side Effects

- `{WRK_HOME}/projects.json` records the home-discovered main via scan.
- Does **not** require or create `~/Projects`.

## Errors

- None on the success path.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("bare --scan-git-repos should default to $HOME and exit 0; got exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolveScanPath(t, req.MainRepo)
	assertStdoutExactPath(t, resp.Stdout, want)
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
}
```
