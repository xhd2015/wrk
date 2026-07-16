## Expected

- Exit code 0 (aborted cleanly, same as plain decline).
- Stdout contains `merge-back aborted`.
- Stdout does **not** contain `synced:`, `tag created`, `tagged `, `pushed main`, propagate vocabulary, or reinstall vocabulary.
- Worktree still exists; main does not have `feature-work`; branch still present.
- Tag `v0.0.2` does **not** exist locally.
- Flag layer must accept `--propagate-tags` and `--reinstall-local` with `--done` (no mutual-exclusion error).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("flag layer still rejects done+propagate-tags+reinstall on abort path; stderr=%q", resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertContains(t, resp.Stdout, "merge-back aborted")
	assertNotContains(t, resp.Stdout, "synced:")
	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "tagged ")
	if strings.Contains(resp.Stdout, "pushed main") {
		t.Fatalf("stdout must not include push after abort; got %q", resp.Stdout)
	}
	assertNoPropagateStageStdout(t, resp.Stdout)
	// Reinstall tail must not run after abort (helpers live under dry-run/; inline here).
	assertNotContains(t, resp.Stdout, "would: go install")
	assertNotContains(t, resp.Stdout, "would: go run")
	assertNotContains(t, resp.Stdout, "would: reinstall ")
	assertNotContains(t, resp.Stdout, "reinstalled ")

	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)

	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not exist when done was aborted")
	}
}
```
