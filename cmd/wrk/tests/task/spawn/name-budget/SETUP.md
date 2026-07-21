# Scenario

**Feature**: fit path basename / branch to ≤255 bytes (reserve 3 for `-99`); shorten slug only after 64-rune soft cap

```
# long repo basename + long task
wrk --task <long> from <long-basename-repo>
  -> create succeeds; filepath.Base(path) and branch byte length ≤ 255
  -> Base == basename + "-" + branch
  -> slug may be shorter than soft-cap slugify output

# prefix alone exceeds budget (no slug room)
  -> clear non-zero error; no silent basename/token chop
```

## Preconditions

- `nameMaxComponentBytes=255`, `nameSuffixReserveBytes=3` → effective max fitted base/branch = 252 before optional `-N`.
- Existing `spawn/long-task` still covers 64-rune soft cap alone.

## Steps

- Helpers build long-basename repos under WorkRoot (each path component ≤255).
- Leaves choose fit-success vs prefix-too-long error.

```go
import (
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	// longRepoBasenameLen forces prefix+64-slug over 255 while prefix alone still fits.
	// path base = {basename}-main-{date}-{slug64}
	// 180 + 1 + 4 + 1 + 10 + 1 + 64 = 261 > 255; prefix without slug = 196 ≤ 252.
	longRepoBasenameLen = 180
	// overBudgetBasenameLen: basename alone with main+date exceeds reserve budget.
	// 240 + 1 + 4 + 1 + 10 = 256 > 252.
	overBudgetBasenameLen = 240
	nameMaxComponentBytes = 255
	nameSuffixReserve     = 3
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	ensureNameBudgetHelpersUsed()
	return nil
}

func longBasename(n int) string {
	return strings.Repeat("r", n)
}

func initLongBasenameRepo(t *testing.T, req *Request, baseLen int) (mainRepo, basename string) {
	t.Helper()
	basename = longBasename(baseLen)
	mainRepo = filepath.Join(req.WorkRoot, basename)
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/longrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	req.RepoDir = mainRepo
	return mainRepo, basename
}

// longTaskDesc yields a task whose soft-cap slug is 64 runes of letters.
func longTaskDesc() string {
	// 80 words-ish of letters/spaces → slug soft-caps at 64 runes after sanitize.
	return "explore the integration of distributed tracing with opentelemetry across all microservices and platforms"
}

func softCapSlug(task string) string {
	return slugify(task)
}

func pathBaseWithoutFit(basename, token, date, slug string) string {
	if slug == "" {
		return basename + "-" + token + "-" + date
	}
	return basename + "-" + token + "-" + date + "-" + slug
}

func branchWithoutFit(token, date, slug string) string {
	if slug == "" {
		return token + "-" + date
	}
	return token + "-" + date + "-" + slug
}

func assertNameBudgetOK(t *testing.T, req *Request, resp *Response, err error, basename, task string) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 under name budget fit, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	wt := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	base := filepath.Base(wt)
	if len(base) > nameMaxComponentBytes {
		t.Fatalf("path basename byte length %d > %d: %q", len(base), nameMaxComponentBytes, base)
	}
	br := gitOutputIsolated(t, wt, "rev-parse", "--abbrev-ref", "HEAD")
	if len(br) > nameMaxComponentBytes {
		t.Fatalf("branch byte length %d > %d: %q", len(br), nameMaxComponentBytes, br)
	}
	// wrk-managed invariant
	wantBase := basename + "-" + br
	if base != wantBase {
		t.Fatalf("invariant Base == basename-branch: base=%q want=%q", base, wantBase)
	}
	// Fixture must require further shorten beyond 64 soft cap when task non-empty.
	if task != "" {
		full := softCapSlug(task)
		if utf8.RuneCountInString(full) < 20 {
			t.Fatalf("fixture slug too short to exercise budget: %q", full)
		}
		unfitted := pathBaseWithoutFit(basename, "main", wrkDate, full)
		if len(unfitted) <= nameMaxComponentBytes {
			t.Fatalf("fixture should overflow 255 without fit: unfitted len=%d base=%q", len(unfitted), unfitted)
		}
		if strings.HasSuffix(base, "-"+full) || strings.HasSuffix(br, "-"+full) {
			t.Fatalf("slug should be shortened below soft-cap for budget; base=%q branch=%q fullSlug=%q", base, br, full)
		}
	}
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, wt, br)
}

func ensureNameBudgetHelpersUsed() {
	_ = longBasename
	_ = initLongBasenameRepo
	_ = longTaskDesc
	_ = softCapSlug
	_ = pathBaseWithoutFit
	_ = branchWithoutFit
	_ = assertNameBudgetOK
}
```
