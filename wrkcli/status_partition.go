package wrkcli

import (
	"sort"

	"github.com/xhd2015/wrk/wrkcli/storage"
)

// StatusPathLists holds ordered primary vs external status paths for main-repo
// status display planning.
type StatusPathLists struct {
	Primary  []string
	External []string
}

// PartitionStatusPaths partitions status paths for main-repo status display planning.
// Primary: mainRoot first, then main-owned linked worktrees in linkedOrdered
// (ListLinked porcelain) order.
// External: scanPaths that are not primary, sorted by normalized path.
// Returned paths are storage.NormalizePath'd.
func PartitionStatusPaths(mainRoot string, scanPaths, linkedOrdered []string) StatusPathLists {
	mainNorm := storage.NormalizePath(mainRoot)

	// Primary membership set (normalized) and ordered primary list.
	primarySet := make(map[string]struct{}, 1+len(linkedOrdered))
	primary := make([]string, 0, 1+len(linkedOrdered))

	// Rule 1: Primary[0] = normalized mainRoot always.
	primary = append(primary, mainNorm)
	primarySet[mainNorm] = struct{}{}

	// Rule 2/4: then each linkedOrdered path in porcelain order; skip main;
	// dedup by normalized path. Paths only in linkedOrdered stay primary.
	for _, p := range linkedOrdered {
		n := storage.NormalizePath(p)
		if _, ok := primarySet[n]; ok {
			continue
		}
		primarySet[n] = struct{}{}
		primary = append(primary, n)
	}

	// Rule 3/5/6: external = scanPaths not already primary; sort lex by norm path.
	extSet := make(map[string]struct{})
	external := make([]string, 0)
	for _, p := range scanPaths {
		n := storage.NormalizePath(p)
		if _, ok := primarySet[n]; ok {
			continue
		}
		if _, ok := extSet[n]; ok {
			continue
		}
		extSet[n] = struct{}{}
		external = append(external, n)
	}
	sort.Strings(external)

	return StatusPathLists{
		Primary:  primary,
		External: external,
	}
}
