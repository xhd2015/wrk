package wrkcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/workops"
)

// tagNextResult is the outcome of a tag-next plan/apply for composition.
type tagNextResult struct {
	// Tags are created tag names (apply) or planned next tag names (dry-run).
	Tags []string
	// MainRepo is the resolved main repository root.
	MainRepo string
}

// runTagNext plans/applies per-scope release tags at HEAD of the resolved main
// repo. Returns created tag names (or planned names on dry-run). The push
// parameter requests tag-only push after apply (legacy tagscope.Apply Push);
// bare --tag-next --push and pipeline composition pass push=false and publish
// branch+tags via runPushMain instead.
func runTagNext(workDir string, dryRun, push, jsonOut bool) ([]string, error) {
	res, err := runTagNextAtResult(workDir, "HEAD", dryRun, push, jsonOut)
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}

// runTagNextAt is like runTagNext but plans/applies tags at headRef (commit or
// symbolic ref). Composition dry-run passes the would-be main tip after a
// planned merge so tag planning is not stuck on stale main HEAD.
func runTagNextAt(workDir, headRef string, dryRun, push, jsonOut bool) ([]string, error) {
	res, err := runTagNextAtResult(workDir, headRef, dryRun, push, jsonOut)
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}

// runTagNextAtResult is the full tag-next outcome for bare tag-next and pipeline
// composition (tags + main repo). Writes human/JSON output to stdout.
func runTagNextAtResult(workDir, headRef string, dryRun, push, jsonOut bool) (tagNextResult, error) {
	return runTagNextAtResultTo(workDir, headRef, dryRun, push, jsonOut, os.Stdout)
}

// runTagNextAtResultTo is like runTagNextAtResult but writes plan/apply output to w.
//
// Core plan/apply is workops.TagNextAll; CLI keeps print/JSON formatting and
// optional tag-only push (legacy push flag).
func runTagNextAtResultTo(workDir, headRef string, dryRun, push, jsonOut bool, w io.Writer) (tagNextResult, error) {
	var out tagNextResult
	if w == nil {
		w = io.Discard
	}
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return out, fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return out, fmt.Errorf("%s is not a git repository", cwd)
	}

	mainRepo, err := resolveMainRepoForWorkDir(cwd)
	if err != nil {
		return out, err
	}
	out.MainRepo = mainRepo

	if headRef == "" {
		headRef = "HEAD"
	}

	plan, collected, err := tagscope.Plan(mainRepo, headRef)
	if err != nil {
		return out, err
	}

	// Core apply / dry-run multi-scope tags via workops.
	core, err := workops.TagNextAll(context.Background(), workops.TagNextOptions{
		Checkout: mainRepo,
		DryRun:   dryRun,
		HeadRef:  headRef,
	})
	if err != nil {
		return out, err
	}
	if core.MainRepo != "" {
		out.MainRepo = core.MainRepo
	}
	out.Tags = core.Tags
	if len(out.Tags) == 0 && dryRun {
		// Fallback to plan names if TagNextAll returned empty without error
		// (should not happen; keeps compose resilient).
		out.Tags = plannedTagNames(plan)
	}

	// Legacy: tagscope.Apply Push=true pushed tag refs only (not branch).
	// Pipeline composition uses push=false + runPushMain for branch+tags.
	if push && !dryRun {
		for _, tag := range out.Tags {
			if tag == "" {
				continue
			}
			if pout, gerr := gitCombinedRunDir(mainRepo, nil, "push", "origin", tag); gerr != nil {
				msg := strings.TrimSpace(string(pout))
				if msg != "" {
					return out, fmt.Errorf("wrk: git push origin %s failed: %s", tag, msg)
				}
				return out, fmt.Errorf("wrk: git push origin %s failed: %w", tag, gerr)
			}
		}
	}

	createdCount := 0
	if !dryRun {
		createdCount = len(out.Tags)
	}

	if jsonOut {
		formatted, err := tagscope.FormatPlanJSON(plan, collected, dryRun, createdCount)
		if err != nil {
			return out, err
		}
		fmt.Fprint(w, formatted)
		return out, nil
	}

	var b strings.Builder
	b.WriteString(tagscope.FormatPlanHuman(plan, collected))
	if !dryRun && len(out.Tags) > 0 {
		tagged, err := tagscope.FormatTaggedLines(mainRepo, headRef, out.Tags)
		if err != nil {
			return out, err
		}
		b.WriteString(tagged)
	}
	b.WriteString(tagscope.FormatPlanSummary(plan, dryRun))
	b.WriteByte('\n')
	fmt.Fprint(w, b.String())
	return out, nil
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
