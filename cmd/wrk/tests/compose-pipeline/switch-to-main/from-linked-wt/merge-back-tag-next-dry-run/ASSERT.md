## Expected

- Exit 0; no confirm noise.
- Primary dry-run: `merge --ff-only <WtBranch>`; **no** `worktree remove` / `branch -D`.
- Tag-next plan after merge (`1 tag planned` / root-bump).
- Order: merge plan before tag plan (activeRoot switch then tag on main).
- Zero mutations: wt kept; no v0.0.2; main HEAD unchanged.

## Side Effects

- None (plan only); worktree remains.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertNoMutexReject(t, resp.Stderr)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertAPNoConfirmNoise(t, resp)

	assertContains(t, resp.Stdout, "merge --ff-only "+req.WtBranch)
	assertNotContains(t, resp.Stdout, "worktree remove")
	assertNotContains(t, resp.Stdout, "branch -D "+req.WtBranch)

	if !strings.Contains(resp.Stdout, "1 tag planned") {
		t.Fatalf("missing tag-next plan after merge-back; stdout=%q", resp.Stdout)
	}
	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	if idxMerge < 0 || idxTag < 0 || idxMerge > idxTag {
		t.Fatalf("want merge plan before tag plan; merge=%d tag=%d\n%s", idxMerge, idxTag, resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "worktree removed:")
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertAPDryRunZeroMutationsLinked(t, req)
}
```
