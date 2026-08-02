# workops Phase 1 — library ops (L2)

## Version
0.0.2

Classic-TDD decision tree for the stable **`workops`** library package
(`github.com/xhd2015/wrk/workops`). Leaves call package APIs **in-process** (L2);
product `wrk` CLI binary is not required. Package largely does not exist yet —
leaves must **RED** until implementer lands symbols.

Phase 1 ops only: **WhereMain**, **Status**, **ListProjects**, **MergeBack**
(dry-run), **TagNext** (dry-run), **Push** (dry-run). Out of scope: spl-cli,
origin→role, fix-vendor, deploy, CodeLens nested `pkgs/shared` tagging, full
unwind Apply.

# DSN (Domain Specific Notion)

- **Caller** — library consumer (tests, later spl) invokes ops with absolute
  checkout / wrk-home paths; no process-global HOME or cwd mutation.
- **workops** — pure library surface over git + wrk storage; injectable paths;
  dry-run modes plan without mutating refs, worktrees, or remotes.
- **WhereMain** — resolve main repository absolute path for a checkout
  (linked worktree → main; main checkout → self, cleaned abs).
- **Status** — structured checkout report: MainPath, CheckoutPath, Branch,
  IsWorktree, HeadShort (dirty counts optional when cheap).
- **ListProjects** — read registered main paths from `{wrkHome}/projects.json`;
  empty wrkHome uses default wrk-home resolution (tests always inject temp home).
- **MergeBack** — land linked worktree into main **without** removing worktree;
  opts WorktreeDir / Sync / DryRun; dry-run must not mutate HEAD or worktree.
- **TagNext** — plan/apply next root release tag at main tip (generic wrk /
  tagscope scheme); DryRun returns planned tag string without creating refs.
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
│       └── ahead-worktree/       # MergeBack DryRun; no mutations
├── tag/
│   └── dry-run/
│       └── root-bump/            # TagNext DryRun → next tag; no tag ref
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
| 5 | land/dry-run/ahead-worktree | MergeBack | DryRun err nil; wt exists; main HEAD unchanged |
| 6 | tag/dry-run/root-bump | TagNext | planned next tag non-empty; no tag created |
| 7 | push/dry-run/with-origin | Push | DryRun err nil; origin branch tip unchanged |

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
	Tag      string // TagNext planned or applied tag
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
		tag, err := workops.TagNext(ctx, workops.TagNextOptions{
			Checkout: req.Checkout,
			DryRun:   req.DryRun,
		})
		resp.Tag = tag
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
//	func TagNext(ctx context.Context, opts TagNextOptions) (tag string, err error)
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
//	type PushOptions struct { Checkout string; DryRun bool; Tags []string }
//
// filepath.Clean / Abs expected on returned main paths.
var _ = filepath.Join
```
