## Expected

- Exit code 0.
- Free-first ship for dirty free leaf:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- **Tag-before-pin (T-tag1 / T-tag2):** combined stdout/stderr shows
  `tag-next example.com/dot-pkgs @ v0.0.2` **before**
  `pin agent-pro <- dot-pkgs @ v0.0.2` (pinReady must not pin planned NextTag
  while free still needs cascade tag-next).
- Mid freeHost after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace for that module
- Mid feature WIP (`FEATURE_WIP.md`) landed when present.
- Combined output must **not** contain `unknown revision` for the free next tag.

## Side Effects

- Free peels early (land/merge-back) before cascade; tag-next is cascade phase.
- Mid freeHost peels early (not pure-consumer deferred); pin of dirty free waits
  for cascade after free tag (not pinReady prelude).
- Top pure pinConsumer peels after cascade pin of mid when dirty.
- Offline file:// modproxy supplies old+next zips so tidy can resolve after tag
  (production VCS fails premature pin with unknown revision; L2 order lock is
  the harness-stable contract).

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		// Today RED: pinReady pins free@next during mid freeHost peel → tidy may
		// fail unknown revision (when next unpublished) or exit non-zero; free
		// tag-next never runs if mid peel aborts first.
		t.Fatalf("T-tag1 three-level freeHost dirty free: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.DepPath == "" {
		t.Fatal("SecondRepo (leaf main) and DepPath (mid main) required")
	}

	// Production symptom surface (must not appear on success).
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("T-tag1: unknown revision after apply (premature pin of untagged free?)\ncombined:\n%s", out)
	}

	// Free tagged and pushed before/with successful pin resolution.
	assertLeafMainAdvancedAndTagged(t, req)

	// Core order lock: cascade free tag-next before mid pin of free @ next.
	assertFreeTagNextBeforeMidPinOfFree(t, out)

	// Mid freeHost pinned to free next; external replace dropped.
	assertMidPinnedToFreeLeaf(t, req)

	// Feature WIP on mid freeHost should land (gen-commit / land) on mid main.
	assertFeatureWIPLanded(t, req.DepPath)
	assertGoModCommittedClean(t, req.DepPath)
}
```
