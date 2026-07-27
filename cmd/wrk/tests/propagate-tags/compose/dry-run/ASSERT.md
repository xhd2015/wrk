
## Expected Output

Two major stages separated by a blank line (propagate uses **planned** next tag):

```
v1.0.0        owned changed                  ->  v1.0.1
1 tag planned

source: /abs/.../lib
  example.com/lib  @ v1.0.1  (tag v1.0.1)

would: update example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.0.1

would: update 1 module across 1 project
```

## Expected

- Exit code 0.
- Stderr empty.
- Tag-next plan footer `1 tag planned` (not `created`).
- Propagate plan shows would-be source release at `v1.0.1` and would-update
  consumer (compose dry-run threads planned tags into propagate).

## Side Effects

- Tag `v1.0.1` does **not** exist.
- Source and app go.mod, HEAD, and tag lists unchanged from pre-run snapshots.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}

	if tagRefExists(t, req.SourcePath, req.NextTag) {
		t.Fatalf("%s tag must not exist after compose dry-run", req.NextTag)
	}
	assertDryRunNoMutation(t, req)

	want := joinMajorStages(
		tagNextRootBumpPlanStdout(req.OldTag, req.NextTag),
		propStageDryRunStdout(req),
	)
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
}
```
