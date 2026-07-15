package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// runTagNext plans/applies per-scope release tags at HEAD of the resolved main
// repo. Returns created tag names (or planned names on dry-run). When push is
// true, tagscope.Apply pushes each new tag to origin (standalone --tag-next --push).
// Callers composing with runPushMain should pass push=false and push tags via
// runPushMain(main, dryRun, tags).
func runTagNext(workDir string, dryRun, push, jsonOut bool) ([]string, error) {
	return runTagNextAt(workDir, "HEAD", dryRun, push, jsonOut)
}

// runTagNextAt is like runTagNext but plans/applies tags at headRef (commit or
// symbolic ref). Composition dry-run passes the would-be main tip after a
// planned merge so tag planning is not stuck on stale main HEAD.
func runTagNextAt(workDir, headRef string, dryRun, push, jsonOut bool) ([]string, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return nil, fmt.Errorf("%s is not a git repository", cwd)
	}

	mainRepo, err := resolveMainRepoForWorkDir(cwd)
	if err != nil {
		return nil, err
	}

	if headRef == "" {
		headRef = "HEAD"
	}

	plan, collected, err := tagscope.Plan(mainRepo, headRef)
	if err != nil {
		return nil, err
	}

	result, err := tagscope.Apply(mainRepo, plan, headRef, tagscope.ApplyOptions{
		DryRun: dryRun,
		Push:   push,
	})
	if err != nil {
		return nil, err
	}

	if jsonOut {
		out, err := tagscope.FormatPlanJSON(plan, collected, dryRun, len(result.Created))
		if err != nil {
			return nil, err
		}
		fmt.Print(out)
		if dryRun {
			return plannedTagNames(plan), nil
		}
		return result.Created, nil
	}

	var b strings.Builder
	b.WriteString(tagscope.FormatPlanHuman(plan, collected))
	if !dryRun && len(result.Created) > 0 {
		tagged, err := tagscope.FormatTaggedLines(mainRepo, headRef, result.Created)
		if err != nil {
			return nil, err
		}
		b.WriteString(tagged)
	}
	b.WriteString(tagscope.FormatPlanSummary(plan, dryRun))
	b.WriteByte('\n')
	fmt.Fprint(os.Stdout, b.String())
	if dryRun {
		// Return planned names so composition --push --dry-run can list would: tag pushes.
		return plannedTagNames(plan), nil
	}
	return result.Created, nil
}

func plannedTagNames(plan tagscope.ChangePlan) []string {
	var tags []string
	for _, d := range plan.Decisions {
		if d.NextTag != "" {
			tags = append(tags, d.NextTag)
		}
	}
	return tags
}
