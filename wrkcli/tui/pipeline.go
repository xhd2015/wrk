package tui

// PlanRecipeStages returns ordered stage IDs included in r for RUN ALL.
// Fixed order matches the dashboard rows (Pre → Main → After).
func PlanRecipeStages(r Recipe) []string {
	var ids []string
	if r.AddAll {
		ids = append(ids, "add-changes")
	}
	if r.GenCommitMsg {
		ids = append(ids, "gen-commit-msg")
	}
	if r.Commit {
		ids = append(ids, "commit")
	}
	if r.MergeBack {
		ids = append(ids, "merge-back")
	}
	if r.Done {
		ids = append(ids, "done")
	}
	if r.Sync {
		ids = append(ids, "sync")
	}
	if r.TagNext {
		ids = append(ids, "tag-next")
	}
	if r.Push {
		ids = append(ids, "push")
	}
	if r.ReinstallLocal {
		ids = append(ids, "reinstall-local")
	}
	return ids
}
