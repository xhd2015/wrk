# Scenario

**Feature**: full pre + primary + posts + reinstall dry-run plans every stage in order; zero mutations (P3)

```
# staged change on linked wtA; wtB behind would-be main; origin; root-bump; GOBIN present stub
myrepo (origin, v0.0.1) + wtA (feature-work ahead + staged) + wtB + gobin/present
  -> wrk --gen-commit-msg --commit --model=m --done --sync --tag-next --push --reinstall-local --dry-run
  -> gen-commit dry plan: mock B / would: git commit (no real commit)
  -> primary MergeBack DryRun plan (ff-merge + remove + branch -D)
  -> blank → would: sync …
  -> blank → tag-next plan v0.0.2
  -> blank → would: git push origin main / v0.0.2
  -> blank → would: go install ./cmd/present + reinstall summary
  -> exit 0; wt still linked; HEAD subject unchanged; staged remains; no new tags; origin unchanged; stub unchanged
```

## Preconditions

- Parent fixtures: `setupDonePipelineSyncWithOrigin`, reinstall seed, dry-run baselines.
- Reuses gen-commit staged-file pattern from `with-gen-commit-msg/` and reinstall seed from `full-combo-reinstall/`.
- Stage order locked: **gen-commit → primary → sync → tag-next → push → reinstall**.

## Steps

1. Root-bump + bare origin + two worktrees; commit ahead on wtA.
2. Stage one additional text file on the worktree (uncommitted index change for gen-commit).
3. Seed reinstall present on main + GOBIN.
4. Snapshot dry-run baseline SHAs + pre-run HEAD subject on the worktree.
5. Run full pre+posts+reinstall with `--dry-run` (no `-y`).

```go
import (
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	setupDonePipelineSyncWithOrigin(t, req)

	// Stage an uncommitted change on the source worktree for gen-commit dry plan.
	staged := filepath.Join(req.WtDir, "staged-for-commit.go")
	writeFile(t, staged, "package staged\n")
	runGitIsolated(t, req.WtDir, "add", "staged-for-commit.go")

	_ = seedDonePipelineReinstallPresent(t, req)
	recordComposeDryRunBaseline(t, req)

	// Record subject before dry-run so Assert can prove no commit.
	subject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	writeFile(t, filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", "wt.head-subject"), subject+"\n")

	req.Args = []string{
		"--gen-commit-msg", "--commit", "--model=m",
		"--done", "--sync", "--tag-next", "--push",
		"--reinstall-local", "--dry-run",
	}
	return nil
}
```
