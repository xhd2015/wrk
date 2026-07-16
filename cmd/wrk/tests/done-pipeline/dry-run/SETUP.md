# Scenario

**Feature**: composition `--dry-run` for `--done` plans optional gen-commit pre + primary (+ cascade + post stages + reinstall tail) with zero mutations and no prompts

```
# P5 + P1 reinstall tail + P2 gen-commit pre: dry-run plans the full requested pipeline;
# never mutates refs / worktrees / remotes / GOBIN / HEAD subject
linked wt (ahead) [+ staged for gen-commit] [+ cascade] [+ origin] [+ post flags] [+ GOBIN stubs]
  -> wrk [--gen-commit-msg --commit] --done --dry-run [--sync] [--tag-next] [--push] [--reinstall-local]
  -> gen-commit (if set): mock B / would: git commit (no real commit)
  -> primary: MergeBack DryRun planned git -C commands (no Proceed?)
  -> cascade (if any): compact would: cascade merge-back <path> per cascaded wt
  -> blank line between major stages
  -> post: would: sync / tag planned / would: git push … (requested only)
  -> reinstall: would: go install|skip: … / would: reinstall N binaries (after other post stages)
  -> wt still linked; no new tags; origin unchanged; no cascade remove; no bin install
```

## Preconditions

- Parent `done-pipeline/` helpers: root-bump seed, bare origin, two-wt sync fixtures,
  `joinMajorStages`, `primaryMergeMsg`, tag helpers, …
- Locked composition dry-run behavior (GREEN for sync/tag/push/reinstall/gen-commit after P1+P2):
  1. `dryRun` is plumbed into `runDone` / cascade / MergeBack / post stages / reinstall tail,
  2. dry-run still **prints** post-stage plans (not skipped solely because Action is dry-run),
  3. post stages plan against **would-be main tip** after planned merge (wt HEAD for ahead/FF),
  4. cascade dry-run does not remove nested linked worktrees,
  5. dry-run never requires confirm TTY / `-y`,
  6. reinstall dry plan scans main tip path; empty/skip-only plan is exit 0,
  7. **P3** full order leaf `full-combo-gen-commit-reinstall/`:
     gen-commit → primary → sync → tag-next → push → reinstall (zero mutations).

## Steps

- Grouping only: leaves set fixtures + `req.Args` (include `--dry-run`; no `-y` required).

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

// recordComposeDryRunBaseline snapshots HEADs / origin for zero-mutation asserts.
func recordComposeDryRunBaseline(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, "_compose_dry_run_baseline")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir baseline: %v", err)
	}
	writeFile(t, filepath.Join(dir, "main.sha"), revParseHEAD(t, req.MainRepo)+"\n")
	writeFile(t, filepath.Join(dir, "wt.sha"), revParseHEAD(t, req.WtDir)+"\n")
	if req.Wt2Dir != "" {
		writeFile(t, filepath.Join(dir, "wt2.sha"), revParseHEAD(t, req.Wt2Dir)+"\n")
	}
	if req.OriginBare != "" {
		writeFile(t, filepath.Join(dir, "origin-main.sha"), revParseRef(t, req.OriginBare, "refs/heads/main")+"\n")
	}
	if req.ExternalWtDir != "" {
		if _, err := os.Stat(req.ExternalWtDir); err == nil {
			writeFile(t, filepath.Join(dir, "external.sha"), revParseHEAD(t, req.ExternalWtDir)+"\n")
		}
	}
	if req.DepsLinkedWtDir != "" {
		if _, err := os.Stat(req.DepsLinkedWtDir); err == nil {
			writeFile(t, filepath.Join(dir, "deps.sha"), revParseHEAD(t, req.DepsLinkedWtDir)+"\n")
		}
	}
}

func readBaselineSHA(t *testing.T, req *Request, name string) string {
	t.Helper()
	p := filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", name)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read baseline %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

// tagNextRootBumpPlanStdout is dry-run tag-next human plan (v0.0.1 → v0.0.2).
// Tip for plan must be would-be main after planned merge (wt HEAD for ahead/FF).
func tagNextRootBumpPlanStdout() string {
	return "v0.0.1        owned changed                  ->  v0.0.2\n1 tag planned\n"
}

func wouldPushMainOrigin(tag string) string {
	if tag == "" {
		return "would: git push origin main\n"
	}
	return fmt.Sprintf("would: git push origin main\nwould: git push origin %s\n", tag)
}

// wouldSyncDistributeOne: post-merge pass-2 would-lines (wt behind planned main tip).
func wouldSyncDistributeOne(branch string, n int) string {
	return fmt.Sprintf(
		"would: %s ← main  (+%d %s)\n\nwould: synced: 0 into main, 1 into worktrees, 0 skipped\n",
		branch, n, syncCommitWord(n),
	)
}

// assertPrimaryMergeBackDryRunPlanned checks MergeBack DryRun planned command listing
// for ahead + remove (done path): ff-merge, worktree remove, branch -D.
// Paths are environment-dependent (pathfmt.Short) so match command shapes only.
func assertPrimaryMergeBackDryRunPlanned(t *testing.T, stdout, wtBranch string) {
	t.Helper()
	if strings.Contains(stdout, "Proceed?") {
		t.Fatalf("dry-run must not prompt; stdout=%q", stdout)
	}
	if strings.Contains(stdout, "stdin is not a terminal") {
		t.Fatalf("dry-run must not require confirm TTY; stderr-like in stdout=%q", stdout)
	}
	// Planned commands (MergeBack printDryRun vocabulary)
	assertContains(t, stdout, "merge --ff-only "+wtBranch)
	assertContains(t, stdout, "worktree remove")
	assertContains(t, stdout, "branch -D "+wtBranch)
	// Must not look like a real apply merge message alone without planning.
	// Real apply prints "merged branch <b> into main" without git -C plan lines.
}

// assertDoneDryRunZeroMutations: wt still linked; main tip unchanged; no new tags;
// origin unchanged when present; optional second wt / cascade paths preserved.
func assertDoneDryRunZeroMutations(t *testing.T, req *Request) {
	t.Helper()
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))

	mainSHA := revParseHEAD(t, req.MainRepo)
	if want := readBaselineSHA(t, req, "main.sha"); mainSHA != want {
		t.Fatalf("main HEAD mutated under dry-run: got %s want %s", mainSHA, want)
	}
	wtSHA := revParseHEAD(t, req.WtDir)
	if want := readBaselineSHA(t, req, "wt.sha"); wtSHA != want {
		t.Fatalf("worktree HEAD mutated under dry-run: got %s want %s", wtSHA, want)
	}
	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not be created under dry-run")
	}
	if req.OriginBare != "" {
		originSHA := revParseRef(t, req.OriginBare, "refs/heads/main")
		if want := readBaselineSHA(t, req, "origin-main.sha"); originSHA != want {
			t.Fatalf("origin/main mutated under dry-run: got %s want %s", originSHA, want)
		}
		if remoteTagExists(t, req.OriginBare, "v0.0.2") {
			t.Fatal("origin must not receive v0.0.2 under dry-run")
		}
	}
	if req.Wt2Dir != "" {
		assertFileExists(t, req.Wt2Dir)
		wt2SHA := revParseHEAD(t, req.Wt2Dir)
		if want := readBaselineSHA(t, req, "wt2.sha"); wt2SHA != want {
			t.Fatalf("second worktree HEAD mutated under dry-run: got %s want %s", wt2SHA, want)
		}
	}
	if req.ExternalWtDir != "" {
		assertFileExists(t, req.ExternalWtDir)
	}
	if req.DepsLinkedWtDir != "" {
		assertFileExists(t, req.DepsLinkedWtDir)
	}
}

// assertNoConfirmPromptNoise: stderr/stdout free of confirm / non-tty prompt errors.
func assertNoConfirmPromptNoise(t *testing.T, resp *Response) {
	t.Helper()
	assertNotContains(t, resp.Stdout, "Proceed?")
	assertNotContains(t, resp.Stderr, "Proceed?")
	assertNotContains(t, resp.Stderr, "stdin is not a terminal")
	assertNotContains(t, resp.Stderr, "cannot prompt")
	assertNotContains(t, resp.Stdout, "cannot prompt")
}

// seedDonePipelineReinstallPresent: add ./cmd/present package main on main tip +
// GOBIN/present stub so composed --reinstall-local dry-run can print would: go install.
// Stores gobin under WorkRoot/gobin and sets ExtraEnv GOBIN=… .
func seedDonePipelineReinstallPresent(t *testing.T, req *Request) string {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("seedDonePipelineReinstallPresent: MainRepo empty")
	}
	// Avoid doctest anti-pattern of "package main" + "func main()" in one literal.
	src := fmt.Sprintf("package %s\n\nfunc main() {}\n", "main")
	cmdDir := filepath.Join(req.MainRepo, "cmd", "present")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("mkdir cmd/present: %v", err)
	}
	writeFile(t, filepath.Join(cmdDir, "main.go"), src)

	binDir := filepath.Join(req.WorkRoot, "gobin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir gobin: %v", err)
	}
	writeFile(t, filepath.Join(binDir, "present"), "stub-binary\n")
	// Ensure executable bit for install target discovery.
	if err := os.Chmod(filepath.Join(binDir, "present"), 0o755); err != nil {
		t.Fatalf("chmod present stub: %v", err)
	}
	req.ExtraEnv = append(req.ExtraEnv, "GOBIN="+binDir)
	return binDir
}

func reinstallPresentDryRunStdout() string {
	return "would: go install ./cmd/present\nwould: reinstall 1 binaries (0 skipped)\n"
}

func assertReinstallDryRunPresent(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "would: go install ./cmd/present") {
		t.Fatalf("stdout missing reinstall would: go install ./cmd/present\nfull:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: reinstall 1 binaries (0 skipped)") {
		t.Fatalf("stdout missing reinstall summary would: reinstall 1 binaries\nfull:\n%s", stdout)
	}
}

func assertStubPresentUnchanged(t *testing.T, binDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(binDir, "present"))
	if err != nil {
		t.Fatalf("read present stub: %v", err)
	}
	if string(data) != "stub-binary\n" {
		t.Fatalf("present bin mutated under dry-run: %q", string(data))
	}
}

func assertNoReinstallStageStdout(t *testing.T, stdout string) {
	t.Helper()
	assertNotContains(t, stdout, "would: go install")
	assertNotContains(t, stdout, "would: go run")
	assertNotContains(t, stdout, "would: reinstall ")
	assertNotContains(t, stdout, "reinstalled ")
	// skip: lines alone are weak; require no reinstall summary vocabulary
}

var (
	_ = recordComposeDryRunBaseline
	_ = readBaselineSHA
	_ = tagNextRootBumpPlanStdout
	_ = wouldPushMainOrigin
	_ = wouldSyncDistributeOne
	_ = assertPrimaryMergeBackDryRunPlanned
	_ = assertDoneDryRunZeroMutations
	_ = assertNoConfirmPromptNoise
	_ = seedDonePipelineReinstallPresent
	_ = reinstallPresentDryRunStdout
	_ = assertReinstallDryRunPresent
	_ = assertStubPresentUnchanged
	_ = assertNoReinstallStageStdout
	_ = fmt.Sprintf
)
```
