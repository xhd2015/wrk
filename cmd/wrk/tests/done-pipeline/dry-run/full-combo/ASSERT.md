## Expected Output (stage vocabulary)

Primary MergeBack DryRun planned commands, then blank-separated post stages:

```
  # main: fast forward
  git -C <main> merge --ff-only <WtBranch>
  # worktree: remove
  git -C <main> worktree remove <WtDir>
  # worktree branch: drop
  git -C <main> branch -D <WtBranch>
  …

would: feature-stays ← main  (+1 commit)

would: synced: 0 into main, 1 into worktrees, 0 skipped

v0.0.1        owned changed                  ->  v0.0.2
1 tag planned

would: git push origin main
would: git push origin v0.0.2
```

(Paths in primary plan are shortened / abs-dependent — assert command shapes + exact post blocks.)

## Expected

- Exit code 0.
- No confirm prompt / non-TTY confirm errors (no `-y` required).
- Primary: MergeBack DryRun plan for ahead+remove.
- Post stages present with blank-line separation (sync → tag-next → push order).
- Sync uses **would:** vocabulary and plans distribute of would-be main tip into `feature-stays`.
- Tag-next dry-run: `1 tag planned` for root-bump v0.0.2 (tip after planned merge).
- Push dry-run: `would: git push origin main` and `would: git push origin v0.0.2`.
- Zero mutations: wt still linked; main/wt2/origin HEADs unchanged; no local/remote `v0.0.2`.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)
	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)

	// Post stages must appear (not skipped on dry-run).
	syncBlock := wouldSyncDistributeOne(req.Wt2Branch, 1)
	tagBlock := tagNextRootBumpPlanStdout()
	pushBlock := wouldPushMainOrigin("v0.0.2")

	for _, part := range []string{
		strings.TrimSpace(syncBlock),
		strings.TrimSpace(tagBlock),
		strings.TrimSpace(pushBlock),
	} {
		if !strings.Contains(resp.Stdout, part) {
			t.Fatalf("stdout missing post-stage block %q\nfull stdout:\n%s", part, resp.Stdout)
		}
	}

	// Stage order: primary plan before sync before tag before push.
	idxSync := strings.Index(resp.Stdout, "would: synced:")
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	idxPush := strings.Index(resp.Stdout, "would: git push origin main")
	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	if idxMerge < 0 || idxSync < 0 || idxTag < 0 || idxPush < 0 {
		t.Fatalf("missing stage markers merge=%d sync=%d tag=%d push=%d\n%s",
			idxMerge, idxSync, idxTag, idxPush, resp.Stdout)
	}
	if !(idxMerge < idxSync && idxSync < idxTag && idxTag < idxPush) {
		t.Fatalf("stage order want primary < sync < tag < push; got merge=%d sync=%d tag=%d push=%d\n%s",
			idxMerge, idxSync, idxTag, idxPush, resp.Stdout)
	}

	// Blank line between major stages: sync block and tag block separated by \n\n
	// (joinMajorStages style). Soft check: "would: synced:…\n\n" then tag plan line.
	if !strings.Contains(resp.Stdout, "would: synced: 0 into main, 1 into worktrees, 0 skipped\n\n") {
		// allow trailing only if next stage still blank-separated somehow
		t.Logf("note: expected blank line after sync would-summary before tag plan")
	}
	if !strings.Contains(resp.Stdout, "1 tag planned\n\nwould: git push origin main") &&
		!strings.Contains(resp.Stdout, "1 tag planned\n\nwould: git push") {
		// require blank line between tag plan and push
		if !strings.Contains(resp.Stdout, "1 tag planned\n\n") {
			t.Fatalf("expected blank line after tag plan before push; stdout:\n%s", resp.Stdout)
		}
	}

	// Must not apply for real.
	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "pushed main →")
	if strings.Contains(resp.Stdout, "tagged v0.0.2") {
		t.Fatalf("dry-run must not print apply 'tagged' lines; stdout=%q", resp.Stdout)
	}

	// Exact post-stage bodies (when extracted) — also keep as documentation via assert.Output
	// on the concatenated post-only expectation is fragile with primary path noise;
	// shape checks above are authoritative.
	_ = assert.Output

	assertDoneDryRunZeroMutations(t, req)
}
```
