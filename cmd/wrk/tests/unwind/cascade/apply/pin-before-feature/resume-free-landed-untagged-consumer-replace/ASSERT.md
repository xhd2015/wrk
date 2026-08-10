## Expected

- Exit code 0.
- Free resume ship:
  - Free main already has next content; cascade creates local tag `v0.0.2` at free main HEAD.
  - Leaf bare origin has `v0.0.2` when push completes (single next tag — no double wrong tags).
- Consumer after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace
  - cascade pin auto-commit present for free @ next
- Combined output shows free `tag-next … @ v0.0.2` and consumer pin of free @ next.
- No no-local-replace hook failure; no `unknown revision`.
- go.mod/go.sum clean on consumer history checkout.
- Replace-only consumer: pin owns the drop (no required feature gen-commit).

## Side Effects

- Free not in dirty peel (already clean on main); cascade still tags owned-changed free.
- Consumer pure pin-consumer deferred until after cascade pin when dirty.
- `--merge-back` keeps linked consumer; `--push` publishes free tag when clear.

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
		t.Fatalf("A5 resume free landed untagged: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}

	assertNoLocalReplaceGenCommitFail(t, out)
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("A5: unknown revision after apply\ncombined:\n%s", out)
	}

	// Free already had next content on main; tag + push complete the resume.
	assertLeafMainAdvancedAndTagged(t, req)
	// Exactly one lightweight next tag on free (idempotent — not v0.0.2 + duplicate).
	tagList := gitOutputIsolated(t, req.SecondRepo, "tag", "--list", unwindApplyNextTag)
	if strings.Count(strings.TrimSpace(tagList), "\n") > 0 {
		// --list one name yields one line when present; empty = missing (caught above).
		_ = tagList
	}
	// Count refs matching next tag name via show-ref (fail if multi-annotated mess).
	showRef := gitOutputIsolated(t, req.SecondRepo, "show-ref", "--tags", "--", "refs/tags/"+unwindApplyNextTag)
	lines := 0
	for _, line := range strings.Split(showRef, "\n") {
		if strings.TrimSpace(line) != "" {
			lines++
		}
	}
	if lines != 1 {
		t.Fatalf("A5: want exactly one refs/tags/%s on free; got %d\n%s",
			unwindApplyNextTag, lines, showRef)
	}

	assertFreeTagNextBeforeConsumerPinOfFree(t, out)
	assertConsumerRequireAndNoExternalReplace(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinCommitPresent(t, hist, req.LeafModulePath, req.ExpectedPinVersion)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertGoModCommittedClean(t, hist)

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
}
```
