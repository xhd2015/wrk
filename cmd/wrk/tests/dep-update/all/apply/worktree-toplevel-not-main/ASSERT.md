## Expected

- Exit 0.
- Apply tree from the linked worktree (`checkout  .`); pin + tidy.
- Summary `updated 1, already 0, skipped 0 in 1 checkouts`.
- Linked worktree go.mod require at v1.2.3.
- **Main** consumer go.mod still at v1.0.0 (not mutated).
- Owner go.mod unchanged.

## Side Effects

- Blast radius is the linked worktree Path (stack member), not MainRepo.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertNotContains(t, resp.Stdout, "would:")
	assertApplyBanner(t, resp.Stdout)
	assertNoArgvDepHeader(t, resp.Stdout)
	assertCheckoutLine(t, resp.Stdout, checkoutLabelOf(req))
	assertPinLine(t, resp.Stdout, modLib, req.WantOldVersion, req.WantVersion)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, wantCheckoutsOf(req), false)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertOwnerGoModUnchanged(t, req)

	// Main checkout must remain on the old require.
	mainGoMod := filepath.Join(req.MainRepo, "go.mod")
	mainBody := readFile(t, mainGoMod)
	hasOld, hasNew := false, false
	for _, line := range strings.Split(mainBody, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace ") {
			continue
		}
		if strings.Contains(trim, modLib) && strings.Contains(trim, "v1.0.0") {
			hasOld = true
		}
		if strings.Contains(trim, modLib) && strings.Contains(trim, "v1.2.3") {
			hasNew = true
		}
	}
	if !hasOld {
		t.Fatalf("main go.mod should still require %s v1.0.0:\n%s", modLib, mainBody)
	}
	if hasNew {
		t.Fatalf("main go.mod must not be bumped to v1.2.3:\n%s", mainBody)
	}
}
```
