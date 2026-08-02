# workops Phase 1 — library ops (L2)

## Version
0.0.2

Classic-TDD decision tree for the stable **`workops`** library package
(`github.com/xhd2015/wrk/workops`). Leaves call package APIs **in-process** (L2);
product `wrk` CLI binary is not required. Expand multi-tag **TagNext** and
**MergeBack Sync** dry-run coverage; existing Phase 1 leaves stay GREEN after
implementer (do not weaken their contracts).

Phase 1 ops: **WhereMain**, **Status**, **ListProjects**, **MergeBack**
(dry-run ± Sync), **TagNext** / **TagNextAll** (dry-run multi-scope), **Push**
(dry-run). Out of scope: spl-cli, origin→role, fix-vendor, deploy, full status
rewrite into workops, full unwind Apply, flaky network apply of Sync.

# DSN (Domain Specific Notion)

- **Caller** — library consumer (tests, later spl / wrkcli rewire) invokes ops
  with absolute checkout / wrk-home paths; no process-global HOME or cwd mutation.
- **workops** — pure library surface over git + wrk storage; injectable paths;
  dry-run modes plan without mutating refs, worktrees, or remotes.
- **WhereMain** — resolve main repository absolute path for a checkout
  (linked worktree → main; main checkout → self, cleaned abs).
- **Status** — structured checkout report: MainPath, CheckoutPath, Branch,
  IsWorktree, HeadShort (dirty counts optional when cheap).
- **ListProjects** — read registered main paths from `{wrkHome}/projects.json`;
  empty wrkHome uses default wrk-home resolution (tests always inject temp home).
- **MergeBack** — land linked worktree into main **without** removing worktree;
  opts WorktreeDir / Sync / DryRun. DryRun never mutates HEAD or worktree.
  When Sync=true under DryRun, still no mutations (post-land sync is planned /
  no-op only). Apply Sync composition is implementer/CLI parity work; L2 leaves
  lock the dry-run safety contract.
- **TagNext / TagNextAll** — plan/apply next release tag(s) at main tip via
  tagscope (root and nested path scopes, e.g. `v0.0.2` and `sub/v0.2.4`).
  **TagNext** returns the primary/first planned-or-created tag string (BC).
  **TagNextAll** returns `TagNextResult{Tags, MainRepo}` with all planned
  (dry-run) or created names — parity with wrkcli `runTagNextAtResult`.
  DryRun creates no tag refs.
- **Push** — push current branch (and optional tags) from checkout/main; DryRun
  is a no-op network (no remote ref change).
- **Fixtures** — per-leaf `t.TempDir()` work roots + `git_isolated` repos /
  linked worktrees; projects.json written under temp wrk home.

## Tree Overview

```
ops/
├── resolve/                      # WhereMain
│   ├── from-linked-worktree/     # wt → main abs ≠ wt
│   └── from-main/                # main → cleaned self
├── status/
│   └── linked-worktree/          # IsWorktree, MainPath, Branch
├── projects/
│   └── registered-path/          # temp wrk home + projects.json entry
├── land/
│   └── dry-run/
│       ├── ahead-worktree/       # MergeBack DryRun Sync=false; no mutations
│       └── with-sync/            # MergeBack DryRun Sync=true; no mutations
├── tag/
│   └── dry-run/
│       ├── root-bump/            # TagNext primary tag v0.0.2; no tag ref
│       └── multi-scope/          # TagNextAll ≥2 planned tags; no refs
└── push/
    └── dry-run/
        └── with-origin/          # Push DryRun; origin tip unchanged
```

## Test Case Index

| # | Leaf | Op | Description |
|---|------|-----|-------------|
| 1 | resolve/from-linked-worktree | WhereMain | linked wt → main abs; path ≠ worktree |
| 2 | resolve/from-main | WhereMain | main checkout → cleaned self path |
| 3 | status/linked-worktree | Status | IsWorktree true; MainPath; Branch non-empty |
| 4 | projects/registered-path | ListProjects | temp wrk home includes registered main path |
| 5 | land/dry-run/ahead-worktree | MergeBack | DryRun Sync=false; wt exists; main HEAD unchanged |
| 6 | land/dry-run/with-sync | MergeBack | DryRun Sync=true; no HEAD/wt mutations |
| 7 | tag/dry-run/root-bump | TagNext | primary planned tag non-empty; no tag created |
| 8 | tag/dry-run/multi-scope | TagNextAll | ≥2 planned tags (root+sub); no tag refs |
| 9 | push/dry-run/with-origin | Push | DryRun err nil; origin branch tip unchanged |

## How to Run

```sh
cd wrk-mad-max
doctest vet ./workops/tests/ops
doctest test ./workops/tests/ops/...
doctest test ./workops/tests/ops/resolve/from-linked-worktree
```

```go
import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/workops"
)

// OpKind selects which workops API Run exercises.
type OpKind string

const (
	OpWhereMain     OpKind = "where-main"
	OpStatus        OpKind = "status"
	OpListProjects  OpKind = "list-projects"
	OpMergeBack     OpKind = "merge-back"
	OpTagNext       OpKind = "tag-next"
	OpPush          OpKind = "push"
)

// Request is filled by root→leaf Setup; paths are absolute (from t.TempDir).
type Request struct {
	Op OpKind

	// Shared fixture layout (set by root Setup / leaf helpers).
	WorkRoot string
	WrkHome  string // temp wrk home; ListProjects injects this explicitly
	MainRepo string // absolute main checkout
	WtDir    string // absolute linked worktree (when used)
	WtBranch string

	// Checkout is the path passed to WhereMain / Status / TagNext / Push.
	Checkout string

	// MergeBack / TagNext / Push options (DryRun always true for P1 dry-run leaves).
	Sync   bool
	DryRun bool
	Tags   []string // optional tag refs for Push

	// Snapshots for dry-run mutation asserts (set in leaf Setup).
	MainHEADBefore   string
	OriginBare       string
	OriginHEADBefore string
}

// Response holds the structured result of the selected op.
type Response struct {
	MainAbs  string
	Status   *workops.StatusReport
	Projects []workops.Project
	// Tag is the primary/first planned-or-created tag (TagNext BC field).
	Tag string
	// Tags is the full multi-scope list from TagNextAll (planned dry-run or created).
	Tags []string
	// TagMainRepo is TagNextResult.MainRepo when TagNextAll is used.
	TagMainRepo string
	// Err is surfaced via Run's error return; Assert also checks side effects.
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	ctx := context.Background()
	resp := &Response{}

	switch req.Op {
	case OpWhereMain:
		mainAbs, err := workops.WhereMain(req.Checkout)
		resp.MainAbs = mainAbs
		return resp, err

	case OpStatus:
		st, err := workops.Status(req.Checkout)
		resp.Status = st
		return resp, err

	case OpListProjects:
		list, err := workops.ListProjects(req.WrkHome)
		resp.Projects = list
		return resp, err

	case OpMergeBack:
		err := workops.MergeBack(ctx, workops.MergeBackOptions{
			WorktreeDir: req.Checkout,
			Sync:        req.Sync,
			DryRun:      req.DryRun,
			WrkHome:     req.WrkHome,
		})
		return resp, err

	case OpTagNext:
		// Prefer TagNextAll so multi-scope leaves can assert all planned names.
		// Populate Tag as the primary (first) entry for root-bump BC asserts.
		// TagNext(string) remains a product BC helper (implementer may wrap TagNextAll).
		res, err := workops.TagNextAll(ctx, workops.TagNextOptions{
			Checkout: req.Checkout,
			DryRun:   req.DryRun,
		})
		resp.Tags = res.Tags
		resp.TagMainRepo = res.MainRepo
		if len(res.Tags) > 0 {
			resp.Tag = res.Tags[0]
		}
		return resp, err

	case OpPush:
		err := workops.Push(ctx, workops.PushOptions{
			Checkout: req.Checkout,
			DryRun:   req.DryRun,
			Tags:     req.Tags,
		})
		return resp, err

	default:
		return nil, fmt.Errorf("unknown op %q", req.Op)
	}
}

// Compile-time shape notes for implementer (not executed here):
//
//	func WhereMain(checkout string) (mainAbs string, err error)
//	func Status(checkout string) (*StatusReport, error)
//	func ListProjects(wrkHome string) ([]Project, error)
//	func MergeBack(ctx context.Context, opts MergeBackOptions) error
//	// Prefer BC: keep TagNext returning primary tag string.
//	func TagNext(ctx context.Context, opts TagNextOptions) (tag string, err error)
//	// Multi-scope parity with wrkcli runTagNextAtResult:
//	func TagNextAll(ctx context.Context, opts TagNextOptions) (TagNextResult, error)
//	func Push(ctx context.Context, opts PushOptions) error
//
// Types (suggested):
//
//	type StatusReport struct {
//	  MainPath, CheckoutPath, Branch, HeadShort string
//	  IsWorktree bool
//	  // Dirty optional: Added, Changed, Renamed, Deleted int
//	}
//	type Project struct { Path string; OriginURL string }
//	type MergeBackOptions struct {
//	  WorktreeDir string; Sync, DryRun bool; WrkHome string
//	}
//	type TagNextOptions struct { Checkout string; DryRun bool }
//	type TagNextResult struct {
//	  Tags     []string // planned (dry-run) or created names, all scopes
//	  MainRepo string   // resolved main absolute path
//	}
//	type PushOptions struct { Checkout string; DryRun bool; Tags []string }
//
// filepath.Clean / Abs expected on returned main paths.
// TagNext may be implemented as first of TagNextAll.Tags (or "" + error).
// MergeBack Sync=true + DryRun=true must not mutate; Sync apply may call
// post-land sync composition (wrk --merge-back --sync parity) in product code.
var _ = filepath.Join
```
