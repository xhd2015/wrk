package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

func runTagNext(workDir string, dryRun, push, jsonOut bool) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	mainRepo, err := resolveMainRepoForWorkDir(cwd)
	if err != nil {
		return err
	}

	plan, collected, err := tagscope.Plan(mainRepo, "HEAD")
	if err != nil {
		return err
	}

	result, err := tagscope.Apply(mainRepo, plan, "HEAD", tagscope.ApplyOptions{
		DryRun: dryRun,
		Push:   push,
	})
	if err != nil {
		return err
	}

	if jsonOut {
		out, err := tagscope.FormatPlanJSON(plan, collected, dryRun, len(result.Created))
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	}

	var b strings.Builder
	b.WriteString(tagscope.FormatPlanHuman(plan, collected))
	if !dryRun && len(result.Created) > 0 {
		tagged, err := tagscope.FormatTaggedLines(mainRepo, "HEAD", result.Created)
		if err != nil {
			return err
		}
		b.WriteString(tagged)
	}
	b.WriteString(tagscope.FormatPlanSummary(plan, dryRun))
	b.WriteByte('\n')
	fmt.Fprint(os.Stdout, b.String())
	return nil
}