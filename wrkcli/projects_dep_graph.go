package wrkcli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// runProjectsDepGraph prints the human-readable cross-project module dependency
// graph for registered WRK_HOME projects. Soft-skips missing registry paths with
// a stderr warning; never requires a git cwd.
func runProjectsDepGraph(wrkHome string, ctx *invocationContext) error {
	inv, err := BuildInventory(wrkHome)
	if err != nil {
		return err
	}
	errw := io.Writer(os.Stderr)
	out := io.Writer(os.Stdout)
	if ctx != nil {
		errw = ctx.errw()
		out = ctx.out()
	}
	for _, p := range inv.SkippedPaths {
		fmt.Fprintf(errw, "warning: project path does not exist: %s\n", p)
	}
	return writeProjectsDepGraph(out, inv)
}

// writeProjectsDepGraph formats Inventory.Projects and CrossEdges to w.
func writeProjectsDepGraph(w io.Writer, inv Inventory) error {
	edges := inv.CrossEdges()
	// Index cross edges by consumer project + consumer module for O(1) lookup.
	type edgeKey struct {
		project string
		module  string
	}
	byConsumer := make(map[edgeKey][]Edge)
	for _, e := range edges {
		k := edgeKey{project: e.ConsumerProject, module: e.ConsumerModule}
		byConsumer[k] = append(byConsumer[k], e)
	}

	var b strings.Builder
	moduleCount := 0
	for i, p := range inv.Projects {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "project %s  (%s)\n", filepath.Base(p.Path), p.Path)
		for _, m := range p.Modules {
			moduleCount++
			dir := m.Dir
			if dir == "" {
				dir = "."
			}
			fmt.Fprintf(&b, "  module %s  dir=%s\n", m.Path, dir)
			for _, e := range byConsumer[edgeKey{project: p.Path, module: m.Path}] {
				fmt.Fprintf(&b, "    → %s@%s  [%s]\n", e.DepPath, e.DepVersion, filepath.Base(e.OwnerProject))
			}
		}
	}
	if len(inv.Projects) > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(formatDepGraphFooter(len(inv.Projects), moduleCount, len(edges)))
	b.WriteByte('\n')
	_, err := io.WriteString(w, b.String())
	return err
}

func formatDepGraphFooter(projects, modules, edges int) string {
	return fmt.Sprintf("%d %s  ·  %d %s  ·  %d %s",
		projects, pluralWord(projects, "project", "projects"),
		modules, pluralWord(modules, "module", "modules"),
		edges, pluralWord(edges, "cross-edge", "cross-edges"),
	)
}

func pluralWord(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
