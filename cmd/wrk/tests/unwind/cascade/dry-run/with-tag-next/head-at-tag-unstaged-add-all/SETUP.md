# Scenario

**Bug**: HEAD still at LatestTag with only unstaged owned change — peeled `--add-all` + gen-commit must tip-plan `tag-next` (willUseAddAllTip)

```
# sole root main: tagged v0.0.1; HEAD == tag; unstaged root.go owned change
# --gen-commit-msg peels --add-all into GenCommitArgs (top-level AddAll false)
root (HEAD@v0.0.1 + unstaged owned WIP)
  -> wrk --unwind --dry-run --gen-commit-msg --commit --add-all --tag-next
  -> would: git add -A / gen-commit-msg / commit
  -> would: tag-next example.com/root @ v0.0.2
  -> exit 0; zero mutations
```

## Steps

1. Seed sole main at LatestTag with unstaged owned change (not committed).
2. Run dry-run with gen-commit + `--add-all` + `--tag-next` (peel path).
3. Expect tip-aware NextTag → `would: tag-next … @ v0.0.2`.

## Context

- **Coverage gap:** `already-main-no-land` commits owned change before run (HEAD
  ahead of tag → tagscope NextTag without tip refresh). `dirty-gomod/with-add-all`
  uses top-level `--add-all` **without** `--gen-commit-msg` (AddAll never peels).
  `consumer-at-latest-wip-must-tag` covers deferred B1 consumer re-tag, not sole-tip
  plan-time tip refresh with peeled `--add-all`.
- Crime scene: unwind compose with unstaged skill edits while HEAD == latest tag
  → `tagged 0` when `willUseAddAllTip` ignored peeled `--add-all`.
- L2 dry-run: no fake agent; asserts plan vocabulary only.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeHeadAtTagUnstagedOwned(t, req)
	req.Args = []string{
		"--unwind", "--dry-run",
		"--gen-commit-msg", "--commit", "--add-all",
		"--tag-next",
	}
	recordUnwindBaseline(t, req)
	return nil
}

// setupCascadeHeadAtTagUnstagedOwned: sole root at LatestTag with unstaged
// owned-file dirt (HEAD still equals the tag). Tip-aware planning needs Mode B
// (--add-all tip) to bump NextTag.
func setupCascadeHeadAtTagUnstagedOwned(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)
	writeGoModRequire(t, mainRepo, unwindRootModule)
	writeFile(t, filepath.Join(mainRepo, "root.go"), "package root\n\nfunc Version() string { return \"old\" }\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "root.go")
	runGitIsolated(t, mainRepo, "commit", "-m", "add root module")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "")

	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.ExpectedPinVersion = unwindApplyNextTag
	req.PeelOrder = []string{"."}

	// Unstaged owned change; HEAD remains at LatestTag (same-commit tagscope).
	writeFile(t, filepath.Join(mainRepo, "root.go"), "package root\n\nfunc Version() string { return \"next\" }\n")
	status := gitOutputIsolated(t, mainRepo, "status", "--porcelain", "--", "root.go")
	if status == "" {
		t.Fatal("expected unstaged root.go after owned WIP")
	}
}
```
