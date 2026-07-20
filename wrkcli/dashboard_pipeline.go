package wrkcli

import (
	"fmt"
	"strings"

	"github.com/xhd2015/wrk/wrkcli/tui"
)

// planDashboardPipelineStages returns ordered stage IDs for a RUN ALL plan.
// Fixed order: Pre (add, gen, commit) → Main (merge-back | done) → After posts.
func planDashboardPipelineStages(r dashboardRecipe) []string {
	return tui.PlanRecipeStages(recipeToTUI(r))
}

func recipeToTUI(r dashboardRecipe) tui.Recipe {
	return tui.Recipe{
		AddAll:         r.addAll,
		GenCommitMsg:   r.genCommitMsg,
		Commit:         r.commit,
		AgentRunner:    r.agentRunner,
		Done:           r.done,
		MergeBack:      r.mergeBack,
		Sync:           r.sync,
		TagNext:        r.tagNext,
		Push:           r.push,
		ReinstallLocal: r.reinstallLocal,
		DryRun:         r.dryRun,
	}
}

// runDashboardPipeline executes the recipe as ordered single-stage mini-composes,
// emitting start/ok/error/skipped StageEvents for the TUI.
//
// log is TUI-safe: lines go to the dashboard Log panel (never fmt/log to the real tty).
// Always writes the full-recipe argv log once (same as compose) so hermetic tests
// that inspect WRK_DASHBOARD_COMPOSE_ARGV_LOG still see the complete plan.
// Mini stages do not overwrite that log.
func runDashboardPipeline(workDir string, r dashboardRecipe, emit func(tui.StageEvent), log tui.LogFunc) error {
	if emit == nil {
		emit = func(tui.StageEvent) {}
	}
	if log == nil {
		log = func(string, string) {}
	}
	// Honor dirt gate: never add when clean.
	if r.addAll && !dashboardHasAddableDirt(workDir) {
		r.addAll = false
	}

	// Full-recipe argv log (once). File path for tests — not a TUI print.
	args := composeArgvFromRecipe(r)
	if err := writeComposeArgvLog(args); err != nil {
		return fmt.Errorf("wrk: write %s: %w", envDashboardComposeArgvLog, err)
	}

	plan := planDashboardPipelineStages(r)
	if len(plan) == 0 {
		log("", "pipeline: empty plan")
		return nil
	}
	log("", "pipeline: "+strings.Join(plan, " → "))

	skipRest := func(from int) {
		for j := from; j < len(plan); j++ {
			emit(tui.StageEvent{StageID: plan[j], Kind: "skipped"})
			log(plan[j], "skipped")
		}
	}

	i := 0
	for i < len(plan) {
		id := plan[i]

		// add-changes: real git add -A (not via compose).
		if id == "add-changes" {
			emit(tui.StageEvent{StageID: id, Kind: "start"})
			log(id, "start: git add -A")
			if err := gitRunDir(workDir, "add", "-A"); err != nil {
				emit(tui.StageEvent{StageID: id, Kind: "error", Err: err.Error()})
				log(id, "error: "+err.Error())
				skipRest(i + 1)
				return err
			}
			emit(tui.StageEvent{StageID: id, Kind: "ok", Result: "staged"})
			log(id, "ok: staged")
			i++
			continue
		}

		// gen-commit-msg and/or commit: one underlying compose (gen+commit flags).
		if id == "gen-commit-msg" || id == "commit" {
			batch := []string{}
			mini := dashboardRecipe{
				agentRunner: r.agentRunner,
				dryRun:      r.dryRun,
			}
			for i < len(plan) && (plan[i] == "gen-commit-msg" || plan[i] == "commit") {
				batch = append(batch, plan[i])
				if plan[i] == "gen-commit-msg" {
					mini.genCommitMsg = true
				}
				if plan[i] == "commit" {
					// --commit only valid with --gen-commit-msg on the argv path.
					mini.genCommitMsg = true
					mini.commit = true
				}
				i++
			}
			for _, s := range batch {
				emit(tui.StageEvent{StageID: s, Kind: "start"})
				log(s, "start")
			}
			// Mini compose may still fmt to process stdout/stderr; TUI capture bridge
			// mirrors those lines into the Log panel. Prefer structured log for status.
			if err := runDashboardComposeWithRecipeOpts(workDir, nil, mini, false); err != nil {
				for _, s := range batch {
					emit(tui.StageEvent{StageID: s, Kind: "error", Err: err.Error()})
					log(s, "error: "+err.Error())
				}
				skipRest(i)
				return err
			}
			// Structured gen-commit result: git log after a real commit; dry-run → planned.
			subject, body, brief := genCommitStagePayload(workDir, mini)
			for _, s := range batch {
				ev := tui.StageEvent{
					StageID: s,
					Kind:    "ok",
					Result:  brief,
					Subject: subject,
					Body:    body,
				}
				// Prefer "committed" brief on the commit stage when we have a real commit.
				if s == "commit" && subject != "" && !mini.dryRun {
					ev.Result = "committed"
				}
				emit(ev)
				if subject != "" && s == "gen-commit-msg" {
					log(s, "ok: msg: "+subject)
				} else if ev.Result != "" {
					log(s, "ok: "+ev.Result)
				} else {
					log(s, "ok")
				}
			}
			continue
		}

		// Main primary without post flags (posts run as separate stages below).
		if id == "merge-back" || id == "done" {
			emit(tui.StageEvent{StageID: id, Kind: "start"})
			log(id, "start")
			mini := dashboardRecipe{dryRun: r.dryRun}
			if id == "done" {
				mini.done = true
			} else {
				mini.mergeBack = true
			}
			if err := runDashboardComposeWithRecipeOpts(workDir, nil, mini, false); err != nil {
				emit(tui.StageEvent{StageID: id, Kind: "error", Err: err.Error()})
				log(id, "error: "+err.Error())
				skipRest(i + 1)
				return err
			}
			emit(tui.StageEvent{StageID: id, Kind: "ok"})
			log(id, "ok")
			i++
			continue
		}

		// After posts: one flag each.
		emit(tui.StageEvent{StageID: id, Kind: "start"})
		log(id, "start")
		mini := dashboardRecipe{dryRun: r.dryRun}
		switch id {
		case "sync":
			mini.sync = true
		case "tag-next":
			mini.tagNext = true
		case "push":
			mini.push = true
		case "reinstall-local":
			mini.reinstallLocal = true
		default:
			emit(tui.StageEvent{StageID: id, Kind: "error", Err: "unknown stage: " + id})
			log(id, "error: unknown stage")
			skipRest(i + 1)
			return fmt.Errorf("wrk: dashboard pipeline: unknown stage %q", id)
		}
		if err := runDashboardComposeWithRecipeOpts(workDir, nil, mini, false); err != nil {
			emit(tui.StageEvent{StageID: id, Kind: "error", Err: err.Error()})
			log(id, "error: "+err.Error())
			skipRest(i + 1)
			return err
		}
		emit(tui.StageEvent{StageID: id, Kind: "ok"})
		log(id, "ok")
		i++
	}
	return nil
}

// genCommitStagePayload builds structured StageEvent fields after a gen/commit mini-compose.
// Real commit → subject/body from git log -1; dry-run → Result "planned"; gen-only → empty.
func genCommitStagePayload(workDir string, mini dashboardRecipe) (subject, body, brief string) {
	if mini.dryRun {
		return "", "", "planned"
	}
	if !mini.commit {
		// Gen-only: message may appear on the capture bridge (stdout); no durable ref.
		return "", "", ""
	}
	subject, body = lastCommitSubjectBody(workDir)
	if subject != "" {
		brief = subject
	}
	return subject, body, brief
}

// lastCommitSubjectBody reads HEAD's subject (%s) and body (%b). Soft-fails to empty.
func lastCommitSubjectBody(workDir string) (subject, body string) {
	s, err := gitOutputDir(workDir, "log", "-1", "--pretty=%s")
	if err != nil {
		return "", ""
	}
	subject = strings.TrimSpace(s)
	b, err := gitOutputDir(workDir, "log", "-1", "--pretty=%b")
	if err != nil {
		return subject, ""
	}
	body = strings.TrimSpace(b)
	return subject, body
}
