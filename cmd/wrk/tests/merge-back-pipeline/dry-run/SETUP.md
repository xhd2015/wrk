# Scenario

**Feature**: composition `--dry-run` for `--merge-back` plans primary + post stages; worktree kept; zero mutations

```
# P5 subset: merge-back dry-run (+ optional tag-next)
linked wt (ahead, root-bump) -> wrk --merge-back --dry-run [--tag-next]
  -> MergeBack DryRun planned commands (ff-merge; **no** worktree remove / branch -D)
  -> blank + tag planned when --tag-next (would-be main tip)
  -> source wt remains; no tags created
```

## Preconditions

- Parent `merge-back-pipeline/` helpers (local root-bump fixture, joinMajorStages, …).
- Locked composition dry-run: `dryRun` is plumbed into `runMergeBack` + post stages with
  would-be tip for tag-next plan (GREEN leaves).

## Steps

- Grouping only.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func recordMergeBackDryRunBaseline(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, "_mb_dry_run_baseline")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir baseline: %v", err)
	}
	writeFile(t, filepath.Join(dir, "main.sha"), revParseHEAD(t, req.MainRepo)+"\n")
	writeFile(t, filepath.Join(dir, "wt.sha"), revParseHEAD(t, req.WtDir)+"\n")
}

func readMBBaselineSHA(t *testing.T, req *Request, name string) string {
	t.Helper()
	p := filepath.Join(req.WorkRoot, "_mb_dry_run_baseline", name)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read baseline %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

func tagNextRootBumpPlanStdoutMB() string {
	return "v0.0.1        owned changed                  ->  v0.0.2\n1 tag planned\n"
}

// assertPrimaryMergeBackKeepDryRunPlanned: ahead + Remove=false → ff-merge only (no remove).
func assertPrimaryMergeBackKeepDryRunPlanned(t *testing.T, stdout, wtBranch string) {
	t.Helper()
	if strings.Contains(stdout, "Proceed?") {
		t.Fatalf("dry-run must not prompt; stdout=%q", stdout)
	}
	assertContains(t, stdout, "merge --ff-only "+wtBranch)
	// Keep path: must NOT plan worktree remove / branch -D.
	assertNotContains(t, stdout, "worktree remove")
	assertNotContains(t, stdout, "branch -D "+wtBranch)
}

func assertMergeBackDryRunZeroMutations(t *testing.T, req *Request) {
	t.Helper()
	assertSourceWorktreeKept(t, req)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
	mainSHA := revParseHEAD(t, req.MainRepo)
	if want := readMBBaselineSHA(t, req, "main.sha"); mainSHA != want {
		t.Fatalf("main HEAD mutated under dry-run: got %s want %s", mainSHA, want)
	}
	wtSHA := revParseHEAD(t, req.WtDir)
	if want := readMBBaselineSHA(t, req, "wt.sha"); wtSHA != want {
		t.Fatalf("worktree HEAD mutated under dry-run: got %s want %s", wtSHA, want)
	}
	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not be created under dry-run")
	}
}

var (
	_ = recordMergeBackDryRunBaseline
	_ = readMBBaselineSHA
	_ = tagNextRootBumpPlanStdoutMB
	_ = assertPrimaryMergeBackKeepDryRunPlanned
	_ = assertMergeBackDryRunZeroMutations
	_ = fmt.Sprintf
)
```
