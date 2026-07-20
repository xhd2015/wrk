package wrkcli

import (
	"github.com/xhd2015/wrk/wrkcli/tui"
)

// runDashboardTeaLoop runs a single Bubble Tea session until CANCEL.
// Interactive UI lives in wrkcli/tui; deps are injected via callbacks.
func runDashboardTeaLoop(workDir string, ctx *invocationContext) error {
	_ = ctx // compose path uses nil ctx + skip via re-entry; events still recorded inside Run
	return tui.RunDashboard(tui.RunDashboardOpts{
		WorkDir:        workDir,
		Status:         "",
		HasAddableDirt: dashboardHasAddableDirt,
		IsMainCheckout: dashboardIsMainCheckout,
		GitAddAll: func(wd string) error {
			return gitRunDir(wd, "add", "-A")
		},
		RunCompose: func(wd string, r tui.Recipe) error {
			return runDashboardComposeWithRecipe(wd, nil, recipeFromTUI(r))
		},
		RunPipeline: func(wd string, r tui.Recipe, emit func(tui.StageEvent), log tui.LogFunc) error {
			return runDashboardPipeline(wd, recipeFromTUI(r), emit, log)
		},
		ComposeArgv: func(r tui.Recipe) []string {
			return composeArgvFromRecipe(recipeFromTUI(r))
		},
		StagePreview: func(wd, stageID string) tui.StagePreviewResult {
			preview, logs := dashboardStagePreview(wd, stageID)
			return tui.StagePreviewResult{Preview: preview, Logs: logs}
		},
	})
}

func recipeFromTUI(r tui.Recipe) dashboardRecipe {
	return dashboardRecipe{
		addAll:         r.AddAll,
		genCommitMsg:   r.GenCommitMsg,
		commit:         r.Commit,
		agentRunner:    r.AgentRunner,
		done:           r.Done,
		mergeBack:      r.MergeBack,
		sync:           r.Sync,
		tagNext:        r.TagNext,
		push:           r.Push,
		reinstallLocal: r.ReinstallLocal,
		dryRun:         r.DryRun,
	}
}
