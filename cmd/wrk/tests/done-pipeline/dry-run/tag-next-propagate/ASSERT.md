## Expected Output (stage vocabulary)

Primary MergeBack DryRun plan, then blank-separated post stages:

```
  # main: fast forward
  git -C <main> merge --ff-only <WtBranch>
  …
v0.0.1        owned changed                  ->  v0.0.2
1 tag planned

source: <MainRepo>
  example.com/lib  @ v0.0.2  (tag v0.0.2)

would: update example.com/app  (project app)
  example.com/lib  v0.0.1 -> v0.0.2

would: update 1 module across 1 project
```

## Expected

- Exit code 0.
- No confirm prompt / non-TTY confirm errors (no `-y` required).
- Post stages present: tag plan **and** would-propagate using **planned** next tag.
- Stage order: primary plan before tag plan before would-update.
- Zero mutations: wt still linked; no local `v0.0.2`; app go.mod/HEAD baseline.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("flag layer still rejects done+propagate dry-run; stderr=%q", resp.Stderr)
	}
	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)

	tagBlock := strings.TrimSpace(tagNextRootBumpPlanStdout())
	propBlock := strings.TrimSpace(propStageDryRunStdout(
		req.MainRepo, req.DepModulePath, pipelinePropOldTag, pipelinePropNextTag,
		filepath.Base(req.SecondRepo),
	))
	if !strings.Contains(resp.Stdout, tagBlock) {
		t.Fatalf("stdout missing tag-next dry-run block %q\nfull:\n%s", tagBlock, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, propBlock) {
		t.Fatalf("stdout missing would-propagate block %q\nfull:\n%s", propBlock, resp.Stdout)
	}

	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	idxProp := strings.Index(resp.Stdout, "would: update "+pipelinePropAppModule)
	if idxMerge < 0 || idxTag < 0 || idxProp < 0 {
		t.Fatalf("missing stage markers merge=%d tag=%d prop=%d\n%s",
			idxMerge, idxTag, idxProp, resp.Stdout)
	}
	if !(idxMerge < idxTag && idxTag < idxProp) {
		t.Fatalf("stage order want primary < tag < propagate; got merge=%d tag=%d prop=%d\n%s",
			idxMerge, idxTag, idxProp, resp.Stdout)
	}

	// Blank line between tag plan and propagate plan.
	if !strings.Contains(resp.Stdout, "1 tag planned\n\n") {
		t.Fatalf("expected blank line after tag plan before propagate; stdout:\n%s", resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "tagged v0.0.2")
	assertNotContains(t, resp.Stdout, "go build ./... ok")
	assertNotContains(t, resp.Stdout, "committed ")

	// Zero mutations on source wt + app.
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	if tagRefExists(t, req.MainRepo, pipelinePropNextTag) {
		t.Fatal("v0.0.2 must not be created under dry-run")
	}
	assertAppUnchangedFromBaseline(t, req)
}
```
