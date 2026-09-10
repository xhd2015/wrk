package unwind

// ExpandLandActions replaces bundled ModeGenCommit nodes with the meta chain
// add-all → gen-commit-msg → commit (optional). Staging is wrk-owned;
// gen-commit-msg is Generate-only. Msg waits on incoming pins (same-repo index
// safety); commit joins msg + pins.
func ExpandLandActions(actions []*Action, addAll, commit bool) []*Action {
	if len(actions) == 0 {
		return actions
	}
	out := make([]*Action, 0, len(actions)+len(actions))
	idMapLast := map[string]string{}
	type pinAfter struct {
		addID  string
		pinIDs []string
	}
	type pinJoin struct {
		lastID string
		pinIDs []string
	}
	var after []pinAfter
	var joins []pinJoin

	for _, a := range actions {
		if a == nil {
			continue
		}
		if a.Mode != ModeGenCommit {
			out = append(out, cloneAction(a))
			continue
		}
		incoming := append([]string(nil), a.Deps...)
		addID := ""
		if addAll {
			add := cloneAction(a)
			add.ID = a.ID + "#add-all"
			add.Mode = ModeAddAll
			add.Detail = ""
			add.Deps = nil
			add.Produces = nil
			add.Consumes = nil
			out = append(out, add)
			addID = add.ID
		}
		if addID != "" && len(incoming) > 0 {
			after = append(after, pinAfter{addID: addID, pinIDs: append([]string(nil), incoming...)})
		}
		msg := cloneAction(a)
		msg.ID = a.ID + "#gen-commit-msg"
		msg.Mode = ModeGenCommitMsg
		msg.Detail = a.Detail
		msgArt := "message:" + a.Lane
		msg.Produces = []string{msgArt}
		// Wait on add-all and on incoming pins. Generating before a same-repo
		// pin races the index: pin's selective commit can leave a message that
		// no longer matches staged WIP (or an empty commit).
		msg.Deps = append([]string(nil), incoming...)
		if addID != "" {
			msg.Deps = append(msg.Deps, addID)
		}
		out = append(out, msg)
		lastID := msg.ID
		if commit {
			c := cloneAction(a)
			c.ID = a.ID + "#commit"
			c.Mode = ModeCommit
			c.Detail = ""
			c.Deps = append([]string{msg.ID}, incoming...)
			c.Consumes = []string{msgArt}
			c.Produces = nil
			out = append(out, c)
			lastID = c.ID
		} else if len(incoming) > 0 {
			joins = append(joins, pinJoin{lastID: lastID, pinIDs: append([]string(nil), incoming...)})
		}
		idMapLast[a.ID] = lastID
	}

	for _, x := range after {
		for _, a := range out {
			if a == nil {
				continue
			}
			found := false
			for _, id := range x.pinIDs {
				if a.ID == id {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			has := false
			for _, d := range a.Deps {
				if d == x.addID {
					has = true
					break
				}
			}
			if !has {
				a.Deps = append(a.Deps, x.addID)
			}
		}
	}

	for _, a := range out {
		if a == nil {
			continue
		}
		mapped := make([]string, 0, len(a.Deps))
		for _, d := range a.Deps {
			if rep, ok := idMapLast[d]; ok {
				mapped = append(mapped, rep)
			} else {
				mapped = append(mapped, d)
			}
		}
		a.Deps = mapped
	}

	for _, x := range joins {
		for _, a := range out {
			if a == nil {
				continue
			}
			waits := false
			for _, d := range a.Deps {
				if d == x.lastID {
					waits = true
					break
				}
			}
			if !waits {
				continue
			}
			for _, id := range x.pinIDs {
				has := false
				for _, d := range a.Deps {
					if d == id {
						has = true
						break
					}
				}
				if !has {
					a.Deps = append(a.Deps, id)
				}
			}
		}
	}
	return out
}

// landMessageArtifactID is the runtime/plan artifact id for a lane's commit message.
func landMessageArtifactID(lane string) string {
	return "message:" + lane
}
