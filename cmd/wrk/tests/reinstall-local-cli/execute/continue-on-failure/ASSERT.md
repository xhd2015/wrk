## Expected Output

```
go install ./cmd/broken
go install ./cmd/good
reinstalled 1, skipped 0, failed 1
```

(Child `go` compiler diagnostics may appear on stderr between progress lines;
stdout wrk-owned lines are the three above when implementer does not mirror
compiler noise onto stdout.)

## Expected

- Exit code **1** (failed > 0).
- Stdout contains progress for both packages in plan order (`broken` then `good`).
- Last content line of stdout is exactly `reinstalled 1, skipped 0, failed 1`.
- No `would:` dry-run vocabulary.
- `$GOBIN/good` is installed (not stub) and prints `good-ok` — prove continue-on-failure.
- Failed count in summary is at least 1 (locked here as exactly 1).

## Side Effects

- `go install ./cmd/broken` fails (compile error); stub for broken may remain or be
  partially written — not asserted.
- `go install ./cmd/good` still runs after the failure and replaces the good stub.

## Errors

- Overall process fails only via exit 1 after finishing the plan (not abort-on-first-error).

## Exit Code

- 1

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitCode(t, resp, 1)
	assertNotContains(t, resp.Stdout, "would:")
	assertContains(t, resp.Stdout, "go install ./cmd/broken")
	assertContains(t, resp.Stdout, "go install ./cmd/good")
	// broken must appear before good (plan sort by bin name)
	iBroken := strings.Index(resp.Stdout, "go install ./cmd/broken")
	iGood := strings.Index(resp.Stdout, "go install ./cmd/good")
	if iBroken < 0 || iGood < 0 || iBroken > iGood {
		t.Fatalf("want broken progress before good; stdout=%q", resp.Stdout)
	}
	assertExecuteSummary(t, resp.Stdout, 1, 0, 1)
	assertBinNotStub(t, req.BinDir, "good")
	assertBinExecutable(t, req.BinDir, "good")
	assertBinRuns(t, req.BinDir, "good", "good-ok")
}
```
