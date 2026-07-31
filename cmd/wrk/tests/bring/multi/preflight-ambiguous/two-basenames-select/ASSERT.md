## Expected Output

```text
Multiple projects match "mydep":
  1) <aaa/mydep>
  2) <zzz/mydep>
Select [1-2]:
Multiple projects match "otherlib":
  1) <aaa/otherlib>
  2) <zzz/otherlib>
Select [1-2]:
will bring:
  mydep → <zzz/mydep>
  otherlib → <aaa/otherlib>
```

## Expected

- Exit code 0.
- Stderr: **two** `Select [` prompts in left→right order (`mydep` then `otherlib`); exactly one listing block per basename.
- Stderr: after both Selects, a multi-only **`will bring:`** plan with one line per arg (`mydep → <abs>`, `otherlib → <abs>`).
- Plan uses raw bring args as display keys and selected absolute paths after `→`.
- Stdout: exactly two lines — external paths for selected mydep then otherlib (order = bring args); no plan text on stdout.
- External worktrees owned by selected dep mains only; replaces for both modules; `/external` gitignore.

## Side Effects

- Preflight prompts once per ambiguous arg; apply reuses resolved abs paths (no third Select).
- `will bring:` is multi-only and appears only after successful full preflight (before create).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	// Reconstruct candidate sets from WorkRoot layout (aaa/zzz + basename).
	mydepA := multiPreflightAbs(t, filepath.Join(req.WorkRoot, "aaa", "mydep"))
	mydepZ := multiPreflightAbs(t, filepath.Join(req.WorkRoot, "zzz", "mydep"))
	otherA := multiPreflightAbs(t, filepath.Join(req.WorkRoot, "aaa", "otherlib"))
	otherZ := multiPreflightAbs(t, filepath.Join(req.WorkRoot, "zzz", "otherlib"))
	mySorted := multiPreflightSorted(t, mydepA, mydepZ)
	otherSorted := multiPreflightSorted(t, otherA, otherZ)

	// Exactly one Select + one listing per ambiguous basename (resolve-once).
	if n := strings.Count(resp.Stderr, "Select ["); n != 2 {
		t.Fatalf("expected exactly two Select prompts, got %d stderr=%q", n, resp.Stderr)
	}
	if n := strings.Count(resp.Stderr, `Multiple projects match "mydep":`); n != 1 {
		t.Fatalf("expected one mydep listing, got %d stderr=%q", n, resp.Stderr)
	}
	if n := strings.Count(resp.Stderr, `Multiple projects match "otherlib":`); n != 1 {
		t.Fatalf("expected one otherlib listing, got %d stderr=%q", n, resp.Stderr)
	}

	// Listings appear left→right: mydep block before otherlib block.
	iMy := strings.Index(resp.Stderr, `Multiple projects match "mydep":`)
	iOther := strings.Index(resp.Stderr, `Multiple projects match "otherlib":`)
	if iMy < 0 || iOther < 0 || iMy > iOther {
		t.Fatalf("expected mydep listing before otherlib listing; stderr=%q", resp.Stderr)
	}

	// Candidate bodies (order / paths).
	assert.Output(t, resp.Stderr, `<contains>
Multiple projects match "mydep":
  1) `+mySorted[0]+`
  2) `+mySorted[1]+`
Select [1-2]:
</contains>`)
	assert.Output(t, resp.Stderr, `<contains>
Multiple projects match "otherlib":
  1) `+otherSorted[0]+`
  2) `+otherSorted[1]+`
Select [1-2]:
</contains>`)

	// will bring: multi plan on stderr after preflight, before apply.
	// Selected: stdin 2\n1\n → mydep=#2 (zzz), otherlib=#1 (aaa).
	selMydep := mySorted[1]
	selOther := otherSorted[0]
	if !strings.Contains(resp.Stderr, "will bring:") {
		t.Fatalf("expected multi-only will bring: plan on stderr; stderr=%q", resp.Stderr)
	}
	// Plan must come after both Select prompts.
	iPlan := strings.Index(resp.Stderr, "will bring:")
	lastSelect := strings.LastIndex(resp.Stderr, "Select [")
	if iPlan < 0 || lastSelect < 0 || iPlan < lastSelect {
		t.Fatalf("will bring: must appear after Select prompts; stderr=%q", resp.Stderr)
	}
	// One line per arg: raw key + → + resolved abs (spacing flexible).
	if !multiPreflightPlanLineHas(resp.Stderr, "mydep", selMydep) {
		t.Fatalf("will bring plan missing mydep → %s; stderr=%q", selMydep, resp.Stderr)
	}
	if !multiPreflightPlanLineHas(resp.Stderr, "otherlib", selOther) {
		t.Fatalf("will bring plan missing otherlib → %s; stderr=%q", selOther, resp.Stderr)
	}
	// Plan is stderr-only (stdout is external paths only).
	assertNotContains(t, resp.Stdout, "will bring:")
	assertNotContains(t, resp.Stdout, "Multiple projects match")

	want1 := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	want2 := bringExternalWorktreePath(req.ConsumerTop, "otherlib", "main", 0)
	req.ExternalWtDir = want1
	req.ExternalWtDir2 = want2
	assertStdoutTwoPathsExact(t, resp.Stdout, want1, want2)

	assertFileExists(t, want1)
	assertFileExists(t, want2)
	assertGitFileIsWorktreeLink(t, want1)
	assertGitFileIsWorktreeLink(t, want2)
	assertWorktreeListContains(t, selMydep, want1)
	assertWorktreeListContains(t, selOther, want2)

	// Unselected candidates (lex #1 mydep, lex #2 otherlib) must not own the worktrees.
	assertWorktreeListNotContains(t, mySorted[0], want1)
	assertWorktreeListNotContains(t, otherSorted[1], want2)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, multiBringDep1Module, want1) {
		t.Fatalf("missing replace %s => %s: %+v", multiBringDep1Module, want1, mod.Replace)
	}
	if !bringHasReplaceForModule(mod, multiBringDep2Module, want2) {
		t.Fatalf("missing replace %s => %s: %+v", multiBringDep2Module, want2, mod.Replace)
	}

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
}
```
