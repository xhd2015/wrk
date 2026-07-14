## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerWt}/external/mydep-main-{WRK_DATE}` where
  `consumerWt` is the linked consumer worktree.
- External path exists as a linked git worktree.
- **The external worktree's `.git` gitdir resolves to the DEP's main repo**
  (`<dep-main>/.git/worktrees/...`), NOT the consumer's main repo
  (`<consumer-main>/.git/worktrees/...`). The dep repo owns the worktree; the
  consumer (whether main or linked worktree) merely hosts it on disk.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	// The dep worktree must be owned by the dep's main repo, not the consumer's
	// main repo. This is the crux of the report: the gitdir was
	// <consumer-main>/.git/worktrees/... instead of <dep-main>/.git/worktrees/...
	depMain := req.DepPath
	gotMain, err := readWorktreeMainRepo(wantPath)
	if err != nil {
		t.Fatalf("read external worktree .git: %v", err)
	}
	if gotMain != depMain {
		t.Fatalf("external worktree .git gitdir resolves to main repo %s, want dep main %s\n"+
			"(the dep worktree is wrongly registered under the consumer's main repo)", gotMain, depMain)
	}

	// The dep repo owns the worktree...
	assertWorktreeListContains(t, depMain, wantPath)
	// ...and neither the consumer's linked worktree nor its main repo own it.
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)
}
```
