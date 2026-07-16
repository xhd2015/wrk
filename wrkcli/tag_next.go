package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// tagNextResult is the outcome of a tag-next plan/apply for composition.
type tagNextResult struct {
	// Tags are created tag names (apply) or planned next tag names (dry-run).
	Tags []string
	// Plan is the full per-scope decision plan (used to thread dry-run next
	// versions into --propagate-tags).
	Plan tagscope.ChangePlan
	// MainRepo is the resolved main repository root.
	MainRepo string
}

// runTagNext plans/applies per-scope release tags at HEAD of the resolved main
// repo. Returns created tag names (or planned names on dry-run). The push
// parameter is passed to tagscope.Apply; bare --tag-next --push and pipeline
// composition pass push=false and publish branch+tags via runPushMain instead.
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

// runTagNextAtResult is the full tag-next outcome including the plan for
// composition with --propagate-tags dry-run (planned next releases).
func runTagNextAtResult(workDir, headRef string, dryRun, push, jsonOut bool) (tagNextResult, error) {
	var out tagNextResult
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
	out.Plan = plan

	result, err := tagscope.Apply(mainRepo, plan, headRef, tagscope.ApplyOptions{
		DryRun: dryRun,
		Push:   push,
	})
	if err != nil {
		return out, err
	}

	if jsonOut {
		formatted, err := tagscope.FormatPlanJSON(plan, collected, dryRun, len(result.Created))
		if err != nil {
			return out, err
		}
		fmt.Print(formatted)
		if dryRun {
			out.Tags = plannedTagNames(plan)
			return out, nil
		}
		out.Tags = result.Created
		return out, nil
	}

	var b strings.Builder
	b.WriteString(tagscope.FormatPlanHuman(plan, collected))
	if !dryRun && len(result.Created) > 0 {
		tagged, err := tagscope.FormatTaggedLines(mainRepo, headRef, result.Created)
		if err != nil {
			return out, err
		}
		b.WriteString(tagged)
	}
	b.WriteString(tagscope.FormatPlanSummary(plan, dryRun))
	b.WriteByte('\n')
	fmt.Fprint(os.Stdout, b.String())
	if dryRun {
		// Return planned names so composition --push --dry-run can list would: tag pushes.
		out.Tags = plannedTagNames(plan)
		return out, nil
	}
	out.Tags = result.Created
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

// runTagNextPropagateCompose runs bare composition:
//
//	tag-next → optional push (branch + tags via runPushMain) → propagate-tags
//
// Flag order is free; stage order is fixed. --json is rejected at the flag layer.
// Matches done-pipeline: tags created locally, then runPushMain for branch+tags.
func runTagNextPropagateCompose(workDir, wrkHome string, dryRun, push bool) error {
	// Create tags locally only; push (if any) is via runPushMain with tag list.
	tagRes, err := runTagNextAtResult(workDir, "HEAD", dryRun, false, false)
	if err != nil {
		return err
	}

	if push {
		fmt.Println() // blank line before push confirmation
		if err := runPushMain(tagRes.MainRepo, dryRun, tagRes.Tags); err != nil {
			return err
		}
	}

	fmt.Println() // blank line between major stages

	var releaseOverride []SourceRelease
	if dryRun {
		// Core compose dry-run contract: plan consumer bumps against planned
		// next versions even though tags do not exist yet.
		releases, err := ResolveSourceReleases(tagRes.MainRepo)
		if err != nil {
			return err
		}
		releaseOverride = applyPlannedTagsToReleases(releases.Releases, tagRes.Plan)
		if len(releaseOverride) == 0 {
			return fmt.Errorf("wrk: no usable release tags for source modules")
		}
	}
	return runPropagateTagsWithReleases(workDir, wrkHome, dryRun, releaseOverride)
}

// applyPlannedTagsToReleases overlays tagscope planned NextTag values onto the
// resolved source release set (by matching tag path-prefix / scope).
func applyPlannedTagsToReleases(releases []SourceRelease, plan tagscope.ChangePlan) []SourceRelease {
	if len(releases) == 0 || len(plan.Decisions) == 0 {
		return releases
	}
	out := make([]SourceRelease, len(releases))
	copy(out, releases)
	for _, d := range plan.Decisions {
		if d.NextTag == "" {
			continue
		}
		version := d.NextTag
		if d.Scope.PathPrefix != "" {
			version = strings.TrimPrefix(d.NextTag, d.Scope.PathPrefix)
		}
		for i := range out {
			if !releaseMatchesTagScope(out[i], d.Scope) {
				continue
			}
			out[i].Tag = d.NextTag
			out[i].Version = version
		}
	}
	return out
}

// releaseMatchesTagScope reports whether a resolved release tag lives in scope.
func releaseMatchesTagScope(r SourceRelease, scope tagscope.TagScope) bool {
	if scope.PathPrefix == "" {
		// Root scope: tags have no path component (e.g. "v1.0.0").
		return !strings.Contains(r.Tag, "/")
	}
	return strings.HasPrefix(r.Tag, scope.PathPrefix)
}
