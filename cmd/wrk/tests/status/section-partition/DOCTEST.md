# wrk status — primary/external path partition helpers (P1)

## Version
0.0.2

Decision tree for **Phase 1** pure helpers that classify status paths into
**primary** (main + linked-to-main) vs **external** (nested/dep scan hits) and
fix print order. Helpers are landed (`wrkcli.PartitionStatusPaths` /
`wrkcli.StatusPathLists`); this tree must stay **GREEN**.

**P2+ (CLI):** `runStatus` primary/external wiring and order-sensitive CLI goldens
live under `status/section-order/` and related `--status` leaves. **P3:** gray
`---- external ----` header when colorEnabled is covered by
`status/color-output/force-color-header` (and `nested-broken-linked/color`) — not
this pure-helper tree.

# DSN (Domain Specific Notion)

- **Partition helper** — pure function in package `wrkcli`:
  `PartitionStatusPaths(mainRoot, scanPaths, linkedOrdered) StatusPathLists`.
  No git, no CLI, no filesystem side effects. Paths are treated as strings;
  membership and order use **normalized** absolute paths
  (`storage.NormalizePath`).
- **mainRoot** — absolute/normalized main checkout path. Always the **first**
  primary entry when statusing main (deduped if also present in `scanPaths` or
  `linkedOrdered`).
- **scanPaths** — paths discovered by `scan_repo.Scan` under main (any input
  order; treated as a set for membership). Nested independent repos /
  dep worktrees under the tree (`external/*`, `task-hub`, …) appear here.
- **linkedOrdered** — `worktree.ListLinked(main)` style path list in **porcelain
  order**. May include in-tree linked WTs, out-of-tree WRK WTs, and
  prunable/dead entries. Order must be preserved in primary after main.
- **Primary membership** — path is `mainRoot`, **or** path appears in
  `linkedOrdered` (prefer ListLinked membership for dead/prunable paths not
  present in scan).
- **Primary order** — `[mainRoot]` then every non-main path from
  `linkedOrdered` in porcelain order, **deduped** by normalized path.
- **External** — paths from `scanPaths` that are **not** primary membership;
  ordered **lexicographically** by normalized absolute path (independent of
  scan input order). Empty when only main ± its linked worktrees.
- **Header** — product later prints `---- external ----` only when external is
  non-empty (plain without color; gray `#90` when colorEnabled). **P1 does not
  print headers**; this tree only asserts ordered path lists.
- **StatusPathLists** — `{ Primary []string; External []string }` both ordered.

## Tree Overview

```
section-partition/
├── no-external/                      # external list empty
│   ├── main-only/                    # scan=[main], linked=[]
│   ├── linked-order/                 # two out-of-tree linked; preserve ListLinked order
│   ├── prunable-linked/              # dead path only in linked → still primary
│   └── scan-dup-linked/              # path in scan + linked → primary once
└── has-external/                     # at least one nested/dep scan path
    ├── main-plus-nested/             # single nested → external=[nested]
    ├── multiple-external-sorted/     # several external; path-sorted
    └── mixed-full/                   # main + in-tree linked + out-of-tree + nesteds
```

Split factor (MECE, significance-first):

1. **External emptiness** (largest outcome impact for later header / sections).
2. Within each branch: concrete membership / order cases from the requirement.

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | no-external/main-only | `scan=[main]`, `linked=[]` → primary=`[main]`, external=`[]` |
| 2 | no-external/linked-order | two out-of-tree linked; primary=`[main, wtZ, wtA]` preserving ListLinked order (not path sort) |
| 3 | no-external/prunable-linked | dead/prunable path only in `linkedOrdered` → still in primary, ListLinked order |
| 4 | no-external/scan-dup-linked | linked path also listed in scan → once in primary only; external empty |
| 5 | has-external/main-plus-nested | `scan=[main, nested]`, `linked=[]` → primary=`[main]`, external=`[nested]` |
| 6 | has-external/multiple-external-sorted | multiple nested/dep paths; external sorted by normalized path regardless of scan order |
| 7 | has-external/mixed-full | main + in-tree linked (also in scan) + out-of-tree linked + nesteds; primary ListLinked order; external nesteds only, sorted; no dups |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/status/section-partition
doctest test ./cmd/wrk/tests/status/section-partition
doctest test ./cmd/wrk/tests/status/section-partition/no-external
doctest test ./cmd/wrk/tests/status/section-partition/has-external
```

Expect **GREEN**: `wrkcli.PartitionStatusPaths` / `wrkcli.StatusPathLists`
implement the semantics above (P1 complete).

```go
import (
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives the pure partition helper.
// Leaves set MainRoot, ScanPaths, LinkedOrdered and the expected Want* lists
// using the same absolute fixture vocabulary (see root SETUP.md).
type Request struct {
	MainRoot      string
	ScanPaths     []string
	LinkedOrdered []string

	// WantPrimary / WantExternal are the expected ordered path lists
	// (raw fixture paths; Assert normalizes before compare).
	WantPrimary  []string
	WantExternal []string
}

type Response struct {
	Primary  []string
	External []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	// Classic TDD: PartitionStatusPaths / StatusPathLists are the production
	// API under design. RED (compile or assert) until implementer lands them.
	lists := wrkcli.PartitionStatusPaths(req.MainRoot, req.ScanPaths, req.LinkedOrdered)
	return &Response{
		Primary:  lists.Primary,
		External: lists.External,
	}, nil
}
```
