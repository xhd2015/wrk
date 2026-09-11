package unwind

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/term"
)

const jobPlanStageTotal = 4

// runUnwindJob is the single apply/dry-run pipeline: snapshot → JobPlan → walk.
func runUnwindJob(workDir, wrkHome string, flags UnwindFlags) error {
	if flags.Color && flags.NoColor {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}
	colorErr := resolveStderrColor(flags.Color, flags.NoColor)
	colorOut := resolveStdoutColor(flags.Color, flags.NoColor)
	st := newStageWriter(jobPlanStageTotal, colorErr, colorOut)

	st.mark(1, "snapshot")
	snap, err := CollectSnapshot(workDir, SnapshotOpts{Cascade: flags.TagNext})
	if err != nil {
		return err
	}
	for _, w := range snap.Inv.Warnings {
		msg := w
		if !strings.HasPrefix(msg, "warning:") && !strings.HasPrefix(msg, "Warning:") {
			msg = "warning: " + msg
		}
		fmt.Fprintln(os.Stderr, paint(msg, ansiOrange, colorErr))
	}
	nRepo, nMod := 0, len(snap.ModuleNodes)
	if snap.Inv.Members != nil {
		nRepo = len(snap.Inv.Members)
	}
	st.detail(fmt.Sprintf("%d repos · %d modules", nRepo, nMod))

	if err := ValidateUnwindFlags(snap.Peel, flags); err != nil {
		return err
	}
	job := BuildJobPlan(snap, flags)
	if len(job.Blockers) > 0 {
		return fmt.Errorf("%s", job.Blockers[0])
	}

	return runJobFromSnapshot(workDir, wrkHome, snap, job, flags, flags.DryRun, st, colorOut)
}

// ApplyUnwindFromSnapshot applies (or dry-runs) the JobPlan for an existing snapshot.
func ApplyUnwindFromSnapshot(workDir, wrkHome string, snap *Snapshot, flags UnwindFlags) error {
	if snap == nil {
		return fmt.Errorf("wrk: no snapshot")
	}
	if err := ValidateUnwindFlags(snap.Peel, flags); err != nil {
		return err
	}
	job := BuildJobPlan(snap, flags)
	if len(job.Blockers) > 0 {
		return fmt.Errorf("%s", job.Blockers[0])
	}
	colorErr := resolveStderrColor(flags.Color, flags.NoColor)
	colorOut := resolveStdoutColor(flags.Color, flags.NoColor)
	st := newStageWriter(jobPlanStageTotal, colorErr, colorOut)
	st.mark(1, "snapshot")
	st.detail("using existing snapshot")
	return runJobFromSnapshot(workDir, wrkHome, snap, job, flags, flags.DryRun, st, colorOut)
}

func runJobFromSnapshot(workDir, wrkHome string, snap *Snapshot, job *JobPlan, flags UnwindFlags, dryRun bool, st *stageWriter, colorOut bool) error {
	byLabel := pickPeelMembersByLabel(snap.Inv.Members)
	checkout := make(map[string]string, len(byLabel))
	for lab, m := range byLabel {
		if m.Path != "" {
			checkout[lab] = m.Path
		} else {
			checkout[lab] = m.MainRepo
		}
	}
	r := &jobRunner{
		workDir:        workDir,
		wrkHome:        wrkHome,
		snap:           snap,
		job:            job,
		flags:          flags,
		dryRun:         dryRun,
		st:             st,
		byLabel:        byLabel,
		checkoutByLane: checkout,
		peeled:         map[string]bool{},
		stats:          &UnwindApplyStats{HadPeels: snap.Peel != nil && len(snap.Peel.PeelOrder) > 0},
	}
	if err := r.run(); err != nil {
		return err
	}
	if !flags.DryRun {
		if line := formatUnwindSummaryLine(*r.stats, flags, colorOut); line != "" {
			fmt.Fprintln(os.Stderr)
			fmt.Println(line)
		}
	}
	return nil
}

func resolveStderrColor(forceColor, noColor bool) bool {
	if noColor {
		return false
	}
	if forceColor {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

type stageWriter struct {
	total    int
	cur      int
	colorErr bool
	colorOut bool
	indent   string
	group    bool
	errW     io.Writer
	outW     io.Writer
}

func newStageWriter(total int, colorErr, colorOut bool) *stageWriter {
	indent := strings.Repeat(" ", len(fmt.Sprintf("[%d/%d] ", total, total)))
	return &stageWriter{
		total: total, colorErr: colorErr, colorOut: colorOut, indent: indent,
		errW: os.Stderr, outW: os.Stdout,
	}
}

func (s *stageWriter) mark(i int, name string) {
	s.cur = i
	s.group = false
	fmt.Fprintf(s.errW, "[%d/%d] %s\n", i, s.total, name)
}

func (s *stageWriter) detail(msg string) {
	fmt.Fprintln(s.errW, s.indent+paint(msg, ansiGrey, s.colorErr))
}

func (s *stageWriter) repo(name string) {
	s.group = true
	fmt.Fprintln(s.outW, s.indent+name)
}

func (s *stageWriter) would(msg string) {
	kind := "would:"
	if s.colorOut {
		kind = paint(kind, ansiGreen, true)
	}
	fmt.Fprintln(s.outW, s.lineIndent()+kind+" "+msg)
}

func (s *stageWriter) ran(msg string) {
	fmt.Fprintln(s.outW, s.lineIndent()+msg)
}

func (s *stageWriter) skipped(why string) {
	fmt.Fprintln(s.outW, s.lineIndent()+"skipped ("+why+")")
}

func (s *stageWriter) lineIndent() string {
	if s.group {
		return s.indent + "   "
	}
	return s.indent
}

type jobRunner struct {
	workDir, wrkHome string
	snap             *Snapshot
	job              *JobPlan
	flags            UnwindFlags
	dryRun           bool
	st               *stageWriter
	byLabel          map[string]StackMember
	checkoutByLane   map[string]string // Path until merge-back/done, then MainRepo
	peeled           map[string]bool
	stats            *UnwindApplyStats
	shipMains        []string
	seenMain         map[string]struct{}

	mu       sync.Mutex
	pathMu   map[string]*sync.Mutex // serialize git ops per checkout/main path
	messages map[string]string      // message:<lane> → commit message
	createdTags map[string][]string // lane → tags created by ModeTagNext this run
	quietRan bool                   // suppress emitRan (progress UI owns status lines)
	prog     *actionProgress
}

func (r *jobRunner) run() error {
	if r.dryRun {
		return r.runDryPhases()
	}
	full := r.fullApplyGraph()
	actions := []*Action{}
	if full != nil {
		actions = full.Actions
	}
	jobs := resolveJobs(r.flags.Jobs, runtime.GOMAXPROCS(0))

	r.st.mark(2, "apply · action graph")
	r.st.detail(fmt.Sprintf("%d actions · jobs=%d", len(actions), jobs))
	if len(actions) == 0 {
		r.st.skipped("no actions")
		r.st.mark(3, "phase-2 · intra-repo update")
		r.st.skipped("included in graph")
		r.st.mark(4, "ship")
		r.st.skipped("included in graph")
		return nil
	}

	if err := r.preflightAutonomous(actions); err != nil {
		return err
	}

	prog := newActionProgress(progressConfig{
		W:      os.Stderr,
		Color:  resolveProgressColor(r.flags.Color, r.flags.NoColor),
		Indent: r.st.indent,
		LaneDisplay: func(lane string) string {
			return r.laneDisplay(lane)
		},
	}, actions)
	r.prog = prog
	prog.Begin(jobs)

	r.messages = map[string]string{}
	r.pathMu = map[string]*sync.Mutex{}
	r.quietRan = true

	err := runActionGraph(context.Background(), actions, jobs, graphExecOpts{
		OnStart: func(a *Action) { prog.Start(a.ID) },
		OnDone:  func(a *Action, err error) { prog.Finish(a.ID, err) },
	}, func(ctx context.Context, a *Action) error {
		return r.applyActionLocked(a)
	})
	prog.Close()
	prog.DumpFailures()
	r.prog = nil
	if err != nil {
		// Durable error after progress lines (not truncated into a status cell).
		return err
	}
	r.st.mark(3, "phase-2 · intra-repo update")
	r.st.detail("included in graph")
	r.st.mark(4, "ship")
	r.st.detail("included in graph")
	return nil
}

// runDryPhases keeps the phase-grouped would: listing for --dry-run readability.
func (r *jobRunner) runDryPhases() error {
	p1, _, _ := splitJobPlanActions(r.job)
	r.st.mark(2, "phase-1 · cross-repo unwind")
	// Always surface inventory peels that have no land actions (follow-local-replace
	// display paths) even when pin/tag actions exist on other lanes.
	peels := r.printPendingPeelsDryRun(p1)
	if len(p1) == 0 {
		if !peels {
			r.st.skipped("no phase-1 actions")
		}
	} else if err := r.runGrouped(p1); err != nil {
		return err
	}
	r.st.mark(3, "phase-2 · intra-repo update")
	r.printPhase2DryRun()
	r.st.mark(4, "ship")
	r.printShipDryRun()
	return nil
}

// printPendingPeelsDryRun emits would: peel <display> for PeelOrder labels that
// have no land actions in phase-1. Returns true if at least one peel was printed.
func (r *jobRunner) printPendingPeelsDryRun(p1 []*Action) bool {
	if r.snap == nil || r.snap.Peel == nil || len(r.snap.Peel.PeelOrder) == 0 {
		return false
	}
	// Skip peels for lanes that already appear as action groups in phase-1.
	hasLane := map[string]bool{}
	for _, a := range p1 {
		if a != nil && a.Lane != "" {
			hasLane[a.Lane] = true
		}
	}
	printed := false
	for _, lab := range r.snap.Peel.PeelOrder {
		if hasLane[lab] {
			continue
		}
		display := lab
		if m, ok := r.byLabel[lab]; ok {
			display = peelDisplayPath(r.workDir, m.Path)
		}
		r.st.would("peel " + display)
		printed = true
	}
	return printed
}

// fullApplyGraph rebuilds the expanded DAG with intact cross-phase deps.
func (r *jobRunner) fullApplyGraph() *ActionGraph {
	if r.snap == nil {
		return nil
	}
	opts := ActionGraphOpts{Cleanup: r.flags.Cleanup}
	return expandLandGraph(BuildActionGraph(r.snap, r.flags, opts), r.flags)
}

// preflightAutonomous refuses apply when merge-back would need a human prompt
// path that cannot run autonomously. Current merge-back uses auto-confirm;
// this still probes DryRun plans so ahead/diverged relations are visible in
// logs and future interactive-only cases can hard-fail here.
func (r *jobRunner) preflightAutonomous(actions []*Action) error {
	landCleans := map[string]bool{} // lane → gen-commit/add-all will clear porcelain first
	for _, a := range actions {
		if a == nil {
			continue
		}
		switch a.Mode {
		case ModeGenCommit, ModeAddAll, ModeGenCommitMsg, ModeCommit:
			if a.Lane != "" {
				landCleans[a.Lane] = true
			}
		}
	}
	for _, a := range actions {
		if a == nil || (a.Mode != ModeMergeBack && a.Mode != ModeDone) {
			continue
		}
		m, ok := r.byLabel[a.Lane]
		if !ok || !m.Linked {
			continue
		}
		// Skip dirty-WT probe when gen-commit/add-all is planned, or when
		// merge-back/done will autoCommitIfDirty before landing.
		if landCleans[a.Lane] || a.Mode == ModeMergeBack || a.Mode == ModeDone {
			continue
		}
		// Probe only: DryRun never mutates. NeedsConfirm (ahead/diverged) is
		// auto-approved at apply time; refuse only when the probe itself errors.
		_, err := worktree.MergeBack(worktree.MergeBackOptions{
			SourcePath: m.Path,
			Remove:     a.Mode == ModeDone,
			DryRun:     true,
			TmpDir:     filepath.Join(r.wrkHome, "worktrees"),
			StashLabel: "wrk-unwind",
			Stdout:     io.Discard,
		})
		if err != nil {
			return fmt.Errorf("wrk: preflight merge-back %s: %w", a.Lane, err)
		}
	}
	return nil
}

func (r *jobRunner) lockPath(path string) func() {
	path = storage.NormalizePath(path)
	if path == "" {
		return func() {}
	}
	r.mu.Lock()
	if r.pathMu == nil {
		r.pathMu = map[string]*sync.Mutex{}
	}
	m := r.pathMu[path]
	if m == nil {
		m = &sync.Mutex{}
		r.pathMu[path] = m
	}
	r.mu.Unlock()
	m.Lock()
	return m.Unlock
}

func (r *jobRunner) applyActionLocked(a *Action) error {
	dir := r.resourcePath(a)
	unlock := r.lockPath(dir)
	defer unlock()
	io := HostIO{}
	if r.prog != nil && a != nil {
		io = r.prog.HostIOFor(a.ID)
	}
	return r.applyAction(a, io)
}

func (r *jobRunner) resourcePath(a *Action) string {
	if a == nil {
		return ""
	}
	switch a.Mode {
	case ModeTagNext, ModePush, ModeSync, ModeReinstall, ModeMergeBack, ModeDone:
		return r.mainOf(a.Lane)
	default:
		return r.checkoutOf(a.Lane)
	}
}

func (r *jobRunner) phaseByRepo(id string) map[string]*ActionGraph {
	if r.job == nil {
		return nil
	}
	if ph := phaseByID(r.job.Phases, id); ph != nil {
		return ph.ByRepo
	}
	return nil
}

func (r *jobRunner) phase2RepoOrder(byRepo map[string]*ActionGraph) []string {
	if len(byRepo) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var order []string
	add := func(repo string) {
		if repo == "" || seen[repo] || byRepo[repo] == nil {
			return
		}
		seen[repo] = true
		order = append(order, repo)
	}
	if r.snap != nil && r.snap.Peel != nil {
		for _, lab := range r.snap.Peel.PeelOrder {
			add(lab)
		}
	}
	rest := make([]string, 0, len(byRepo))
	for repo := range byRepo {
		if !seen[repo] {
			rest = append(rest, repo)
		}
	}
	sort.Strings(rest)
	return append(order, rest...)
}

func (r *jobRunner) printPhase2DryRun() {
	byRepo := r.phaseByRepo("modules")
	order := r.phase2RepoOrder(byRepo)
	if len(order) == 0 {
		r.st.skipped("no actions")
		return
	}
	for _, repo := range order {
		r.st.repo(r.laneDisplay(repo))
		g := byRepo[repo]
		if g == nil || len(g.Actions) == 0 {
			note := phase2EmptyNote
			if g != nil && g.FilterNote != "" {
				note = g.FilterNote
			}
			r.st.skipped(note)
			continue
		}
		for _, a := range topoActions(g.Actions) {
			r.printWould(a)
		}
	}
}

func (r *jobRunner) printShipDryRun() {
	byRepo := r.phaseByRepo("ship")
	order := r.phase2RepoOrder(byRepo)
	any := false
	for _, repo := range order {
		g := byRepo[repo]
		if g == nil || len(g.Actions) == 0 {
			continue
		}
		any = true
		r.st.repo(r.laneDisplay(repo))
		for _, a := range topoActions(g.Actions) {
			r.printWould(a)
		}
	}
	if !any {
		r.st.skipped("no ship actions")
	}
}

func (r *jobRunner) applyPhase2() error {
	byRepo := r.phaseByRepo("modules")
	order := r.phase2RepoOrder(byRepo)
	if len(order) == 0 {
		r.st.skipped("no actions")
		return nil
	}
	for _, repo := range order {
		r.st.repo(r.laneDisplay(repo))
		g := byRepo[repo]
		if g == nil || len(g.Actions) == 0 {
			note := phase2EmptyNote
			if g != nil && g.FilterNote != "" {
				note = g.FilterNote
			}
			r.st.skipped(note)
			continue
		}
		for _, a := range topoActions(g.Actions) {
			if err := r.applyAction(a, HostIO{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *jobRunner) applyShipPhase() error {
	byRepo := r.phaseByRepo("ship")
	order := r.phase2RepoOrder(byRepo)
	any := false
	for _, repo := range order {
		g := byRepo[repo]
		if g == nil || len(g.Actions) == 0 {
			continue
		}
		any = true
		r.st.repo(r.laneDisplay(repo))
		for _, a := range topoActions(g.Actions) {
			if err := r.applyAction(a, HostIO{}); err != nil {
				return err
			}
		}
	}
	if !any {
		r.st.skipped("no ship actions")
	}
	return nil
}

func actionRepo(a *Action) string {
	if a == nil {
		return ""
	}
	if a.Subject.RepoLabel != "" {
		return a.Subject.RepoLabel
	}
	return a.Lane
}

func (r *jobRunner) runGrouped(actions []*Action) error {
	byRepo := map[string][]*Action{}
	for _, a := range actions {
		if a == nil {
			continue
		}
		byRepo[actionRepo(a)] = append(byRepo[actionRepo(a)], a)
	}
	seen := map[string]bool{}
	var order []string
	for _, a := range actions {
		repo := actionRepo(a)
		if repo == "" || seen[repo] {
			continue
		}
		seen[repo] = true
		order = append(order, repo)
	}
	if len(order) == 0 {
		r.st.skipped("no actions")
		return nil
	}
	for _, repo := range order {
		r.st.repo(r.laneDisplay(repo))
		for _, a := range byRepo[repo] {
			if r.dryRun {
				r.printWould(a)
				continue
			}
			if err := r.applyAction(a, HostIO{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func splitJobPlanActions(job *JobPlan) (phase1, phase2, ship []*Action) {
	if job == nil {
		return nil, nil, nil
	}
	for _, ph := range job.Phases {
		switch ph.ID {
		case "snapshot":
			continue
		case "repos":
			if ph.ActionGraph == nil {
				continue
			}
			phase1 = append(phase1, ph.ActionGraph.Actions...)
		case "modules":
			repos := make([]string, 0, len(ph.ByRepo))
			for repo := range ph.ByRepo {
				repos = append(repos, repo)
			}
			sort.Strings(repos)
			for _, repo := range repos {
				g := ph.ByRepo[repo]
				if g == nil {
					continue
				}
				phase2 = append(phase2, g.Actions...)
			}
		case "ship":
			repos := make([]string, 0, len(ph.ByRepo))
			for repo := range ph.ByRepo {
				repos = append(repos, repo)
			}
			sort.Strings(repos)
			for _, repo := range repos {
				g := ph.ByRepo[repo]
				if g == nil {
					continue
				}
				ship = append(ship, g.Actions...)
			}
		}
	}
	return topoActions(phase1), topoActions(phase2), topoActions(ship)
}

func isShipMode(m ActionMode) bool {
	return m == ModePush || m == ModeSync
}

func topoActions(actions []*Action) []*Action {
	byID := make(map[string]*Action, len(actions))
	idx := make(map[string]int, len(actions))
	for i, a := range actions {
		if a == nil || a.ID == "" {
			continue
		}
		byID[a.ID] = a
		idx[a.ID] = i
	}
	var out []*Action
	seen := map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		a := byID[id]
		if a == nil || seen[id] {
			return
		}
		seen[id] = true
		deps := append([]string(nil), a.Deps...)
		sort.Slice(deps, func(i, j int) bool { return idx[deps[i]] < idx[deps[j]] })
		for _, d := range deps {
			visit(d)
		}
		out = append(out, a)
	}
	ids := make([]string, 0, len(actions))
	for _, a := range actions {
		if a != nil && a.ID != "" {
			ids = append(ids, a.ID)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool { return idx[ids[i]] < idx[ids[j]] })
	for _, id := range ids {
		visit(id)
	}
	return out
}

func (r *jobRunner) printWould(a *Action) {
	for _, line := range r.actionLines(a) {
		r.st.would(line)
	}
}

func (r *jobRunner) emitRan(a *Action) {
	if r.quietRan {
		return
	}
	for _, line := range r.actionLines(a) {
		r.st.ran(line)
	}
}

func (r *jobRunner) actionLines(a *Action) []string {
	if a == nil {
		return nil
	}
	switch a.Mode {
	case ModeGenCommit:
		var out []string
		if r.flags.AddAll || genArgsHasFlag(r.flags.GenCommitArgs, "--add-all") {
			out = append(out, "git add -A")
		}
		out = append(out, "gen-commit-msg")
		if genArgsHasFlag(r.flags.GenCommitArgs, "--commit") {
			out = append(out, "commit")
		}
		return out
	case ModeAddAll:
		return []string{"git add -A"}
	case ModeGenCommitMsg:
		return []string{"gen-commit-msg"}
	case ModeMergeBack:
		return []string{"merge-back"}
	case ModeDone:
		return []string{"done"}
	case ModeTagNext:
		return []string{"tag-next " + a.Subject.Display + " @ " + a.Detail}
	case ModeDepUpdate, ModePin:
		from, to := r.bumpVersions(a)
		msg := a.Subject.Display + " <- " + a.DepModule
		if from != "" && to != "" {
			msg += " " + from + " -> " + to
		} else if a.Detail != "" {
			msg += " " + a.Detail
		}
		return []string{"dep-update " + msg, "go mod tidy  (local git)"}
	case ModeCommit:
		if len(a.Consumes) > 0 {
			return []string{"commit"}
		}
		if a.Detail != "" {
			return []string{"git commit -m '" + a.Detail + "'"}
		}
		return []string{"commit"}
	case ModePush:
		return []string{"push"}
	case ModeSync:
		return []string{"sync"}
	case ModeReinstall:
		return []string{"reinstall-local"}
	default:
		return []string{string(a.Mode) + " " + a.Subject.Display}
	}
}

func (r *jobRunner) laneDisplay(lane string) string {
	if m, ok := r.byLabel[lane]; ok {
		return peelDisplayPath(r.workDir, m.Path)
	}
	return lane
}

func (r *jobRunner) bumpVersions(a *Action) (from, to string) {
	to = goRequireVersionFromTag(a.PinVersion)
	if r.snap == nil {
		return "", to
	}
	for _, n := range r.snap.ModuleNodes {
		if n.Path == a.DepModule {
			from = goRequireVersionFromTag(n.LatestTag)
			if to == "" {
				to = goRequireVersionFromTag(n.NextTag)
			}
			return from, to
		}
	}
	return from, to
}

func (r *jobRunner) applyAction(a *Action, io HostIO) error {
	if a == nil {
		return nil
	}
	// Phase-2 compose commit cards are display-only (pin already committed).
	if a.Mode == ModeCommit && len(a.Consumes) == 0 && a.Reason == ReasonPropagate {
		return nil
	}
	r.emitRan(a)
	switch a.Mode {
	case ModeGenCommit:
		return r.applyGenCommit(a.Lane, io)
	case ModeAddAll:
		return r.applyAddAll(a.Lane)
	case ModeGenCommitMsg:
		return r.applyGenCommitMsgOnly(a, io)
	case ModeMergeBack:
		return r.applyMergeBack(a.Lane, false, io)
	case ModeDone:
		return r.applyMergeBack(a.Lane, true, io)
	case ModeTagNext:
		return r.applyTag(a)
	case ModeDepUpdate, ModePin:
		return r.applyPin(a)
	case ModeCommit:
		return r.applyLandCommit(a, io)
	case ModePush, ModeSync, ModeReinstall:
		return r.applyShip(a, io)
	default:
		return nil
	}
}

func (r *jobRunner) applyAddAll(label string) error {
	dir := r.checkoutOf(label)
	if dir == "" {
		return fmt.Errorf("wrk: add-all %s: no checkout", label)
	}
	if err := gitRunDir(dir, "add", "-A"); err != nil {
		return fmt.Errorf("wrk: git add -A in %s: %w", dir, err)
	}
	return nil
}

func (r *jobRunner) applyGenCommitMsgOnly(a *Action, io HostIO) error {
	label := a.Lane
	dir := r.checkoutOf(label)
	if dir == "" {
		return fmt.Errorf("wrk: gen-commit-msg %s: no checkout", label)
	}
	// Strip staging/commit flags — those are separate meta nodes.
	genArgs := stripGenCommitBoolFlags(r.flags.GenCommitArgs)
	msg, err := runGenerateCommitMsg(dir, genArgs, io)
	if err != nil {
		if isNoStagedCommitErr(err) || strings.Contains(err.Error(), "no staged") {
			return nil
		}
		return err
	}
	art := landMessageArtifactID(label)
	r.mu.Lock()
	if r.messages == nil {
		r.messages = map[string]string{}
	}
	r.messages[art] = msg
	r.mu.Unlock()
	return nil
}

func (r *jobRunner) applyLandCommit(a *Action, io HostIO) error {
	if len(a.Consumes) == 0 {
		return nil
	}
	label := a.Lane
	dir := r.checkoutOf(label)
	if dir == "" {
		return fmt.Errorf("wrk: commit %s: no checkout", label)
	}
	staged, err := gitOutputDir(dir, "diff", "--cached", "--name-only")
	if err != nil {
		return fmt.Errorf("wrk: commit %s: check staged: %w", label, err)
	}
	if strings.TrimSpace(staged) == "" {
		// Pin may have already committed go.mod via --only; feature WIP may be
		// empty. Soft-skip like legacy allowEmptySkip rather than failing apply.
		// With --add-all, commit leftover surgical go.mod/go.sum dirt from
		// partial-edit so pin-only consumers do not leave porcelain.
		if r.flags.AddAll || genArgsHasFlag(r.flags.GenCommitArgs, "--add-all") {
			return autoCommitIfDirty(dir)
		}
		return nil
	}
	art := a.Consumes[0]
	r.mu.Lock()
	msg := r.messages[art]
	r.mu.Unlock()
	if strings.TrimSpace(msg) == "" {
		// Empty message after soft-skip gen-commit-msg: try auto-commit if dirty.
		return autoCommitIfDirty(dir)
	}
	noVerify := genArgsHasFlag(r.flags.GenCommitArgs, "--no-verify")
	return runGitCommitWithMsg(dir, msg, noVerify, io)
}

func (r *jobRunner) checkoutOf(label string) string {
	if label != "" && r.checkoutByLane != nil {
		if p := r.checkoutByLane[label]; p != "" {
			return p
		}
	}
	m := r.byLabel[label]
	if m.Path != "" {
		return m.Path
	}
	return m.MainRepo
}

func (r *jobRunner) mainOf(label string) string {
	m := r.byLabel[label]
	if m.MainRepo != "" {
		return m.MainRepo
	}
	return m.Path
}

func (r *jobRunner) applyGenCommit(label string, io HostIO) error {
	dir := r.checkoutOf(label)
	if dir == "" {
		return fmt.Errorf("wrk: gen-commit %s: no checkout", label)
	}
	if r.flags.AddAll || genArgsHasFlag(r.flags.GenCommitArgs, "--add-all") {
		if err := gitRunDir(dir, "add", "-A"); err != nil {
			return fmt.Errorf("wrk: git add -A in %s: %w", dir, err)
		}
	}
	if err := runGenCommitMsgStage(dir, r.flags.GenCommitArgs, false, false, io); err != nil {
		if !isNoStagedCommitErr(err) {
			return err
		}
		if err := autoCommitIfDirty(dir); err != nil {
			return err
		}
	}
	return nil
}

func (r *jobRunner) applyMergeBack(label string, remove bool, io HostIO) error {
	m, ok := r.byLabel[label]
	if !ok {
		return fmt.Errorf("wrk: merge-back %s: unknown lane", label)
	}
	if !m.Linked {
		r.checkoutByLane[label] = r.mainOf(label)
		return nil
	}
	// --done/--merge-back without gen-commit: stage+commit tracked porcelain so
	// MergeBack can run. Skip when gen-commit already ran (must not scoop
	// untracked leftovers that --add-all did not request).
	if !r.flags.GenCommitMsg {
		if err := autoCommitIfDirty(m.Path); err != nil {
			return err
		}
	}
	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: m.Path,
		TargetPath: "",
		Remove:     remove,
		DryRun:     false,
		TmpDir:     filepath.Join(r.wrkHome, "worktrees"),
		StashLabel: "wrk-unwind",
		Stdout:     io.Out(),
		// Autonomous: never prompt (assume yes for NeedsConfirm plans).
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return true, nil
		},
	})
	if err != nil {
		return mapMergeBackSharedError(err, "--unwind")
	}
	if result.Action == "aborted" {
		return fmt.Errorf("wrk: merge-back aborted during unwind %s", label)
	}
	if result.Message != "" {
		fmt.Fprintln(io.Out(), result.Message)
	}
	main := result.TargetPath
	if main == "" {
		main = r.mainOf(label)
	}
	r.checkoutByLane[label] = storage.NormalizePath(main)
	r.peeled[label] = true
	r.addMain(main)
	if r.stats != nil {
		r.stats.Peeled++
	}
	// Pure pin-consumers often had empty NextTag at plan time (HEAD==LatestTag +
	// WIP only). After feature land, tip is ahead — catch up tagscope next tag.
	if err := r.catchUpTagsAfterLand(label, main); err != nil {
		return err
	}
	return nil
}

// catchUpTagsAfterLand creates next release tags for lane modules whose main HEAD
// advanced past LatestTag when no tag-next action was planned for that module.
func (r *jobRunner) catchUpTagsAfterLand(label, main string) error {
	if !r.flags.TagNext || r.snap == nil || main == "" || label == "" {
		return nil
	}
	planned := map[string]bool{}
	if g := r.fullApplyGraph(); g != nil {
		for _, a := range g.Actions {
			if a != nil && a.Mode == ModeTagNext && a.Lane == label && a.Subject.Module != "" {
				planned[a.Subject.Module] = true
			}
		}
	}
	main = storage.NormalizePath(main)
	head, err := gitOutputDir(main, "rev-parse", "HEAD")
	if err != nil {
		return nil
	}
	head = strings.TrimSpace(head)
	for _, n := range r.snap.ModuleNodes {
		if n.RepoLabel != label || n.LatestTag == "" || n.Path == "" || planned[n.Path] {
			continue
		}
		tagCommit, terr := gitOutputDir(main, "rev-parse", n.LatestTag+"^{commit}")
		if terr != nil {
			tagCommit, terr = gitOutputDir(main, "rev-parse", n.LatestTag)
		}
		if terr != nil || strings.TrimSpace(tagCommit) == "" || strings.TrimSpace(tagCommit) == head {
			continue
		}
		next, ierr := tagscope.IncrementTag(n.LatestTag)
		if ierr != nil || next == "" {
			continue
		}
		if _, verr := gitOutputDir(main, "rev-parse", "--verify", "--quiet", "refs/tags/"+next); verr == nil {
			continue
		}
		if err := cascadeCreateOneTag(main, next); err != nil {
			if _, verr := gitOutputDir(main, "rev-parse", "--verify", "--quiet", "refs/tags/"+next); verr == nil {
				continue
			}
			return err
		}
		r.mu.Lock()
		if r.createdTags == nil {
			r.createdTags = map[string][]string{}
		}
		r.createdTags[label] = append(r.createdTags[label], next)
		r.mu.Unlock()
		if r.stats != nil {
			r.stats.Tagged++
		}
	}
	return nil
}

func (r *jobRunner) applyTag(a *Action) error {
	tag := a.Detail
	if tag == "" {
		return nil
	}
	m, ok := r.byLabel[a.Lane]
	if !ok {
		return fmt.Errorf("wrk: tag-next %s: unknown lane %s", a.Subject.Display, a.Lane)
	}
	main := m.MainRepo
	if main == "" {
		main = m.Path
	}
	if err := cascadeCreateOneTag(main, tag); err != nil {
		// Idempotent: catch-up or a prior action may have created the tag.
		if _, verr := gitOutputDir(main, "rev-parse", "--verify", "--quiet", "refs/tags/"+tag); verr != nil {
			return err
		}
	}
	r.mu.Lock()
	if r.createdTags == nil {
		r.createdTags = map[string][]string{}
	}
	r.createdTags[a.Lane] = append(r.createdTags[a.Lane], tag)
	r.mu.Unlock()
	if r.stats != nil {
		r.stats.Tagged++
	}
	r.addMain(main)
	return nil
}

// tagsForLane returns release tags to publish with push for this lane:
// tags created this run, else planned tag-next Detail from the full graph.
func (r *jobRunner) tagsForLane(lane string) []string {
	r.mu.Lock()
	created := append([]string(nil), r.createdTags[lane]...)
	r.mu.Unlock()
	if len(created) > 0 {
		return created
	}
	g := r.fullApplyGraph()
	if g == nil {
		return nil
	}
	seen := map[string]bool{}
	var tags []string
	for _, a := range g.Actions {
		if a == nil || a.Mode != ModeTagNext || a.Lane != lane || a.Detail == "" {
			continue
		}
		if seen[a.Detail] {
			continue
		}
		seen[a.Detail] = true
		tags = append(tags, a.Detail)
	}
	return tags
}

func (r *jobRunner) applyPin(a *Action) error {
	if a.Subject.Module == "" || a.DepModule == "" || a.PinVersion == "" {
		return nil
	}
	nodes := map[string]UnwindGraphModuleNode{}
	if r.snap != nil {
		for _, n := range r.snap.ModuleNodes {
			if n.Path != "" {
				nodes[n.Path] = n
			}
		}
	}
	cons, ok := nodes[a.Subject.Module]
	if !ok {
		return fmt.Errorf("wrk: dep-update: unknown consumer %s", a.Subject.Module)
	}
	dep := nodes[a.DepModule]
	m, ok := r.byLabel[cons.RepoLabel]
	if !ok {
		return fmt.Errorf("wrk: dep-update %s: no stack member", cons.RepoLabel)
	}
	// Prefer inventory checkout, but pin clean linked Path on MainRepo
	// (C-RI3 nested cmd←parent; reinstall useMain sees pin+tidy).
	checkout := cascadePinCheckout(m)
	if a.Mode == ModePin && m.Linked && m.MainRepo != "" && m.Path != "" {
		if err := worktree.IsClean(m.Path); err == nil {
			checkout = storage.NormalizePath(m.MainRepo)
		}
	}
	if checkout == "" {
		checkout = r.checkoutOf(cons.RepoLabel)
	}
	if checkout == "" {
		checkout = m.MainRepo
	}
	modDir := checkout
	if cons.Dir != "" && cons.Dir != "." {
		modDir = filepath.Join(checkout, filepath.FromSlash(cons.Dir))
	}
	ver := a.PinVersion
	depDir := ""
	if dm, ok := r.byLabel[dep.RepoLabel]; ok {
		depDir = dm.MainRepo
		if depDir == "" {
			depDir = dm.Path
		}
	}
	// Partial-edit when go.mod/go.sum dirty: pin on Base, selective commit,
	// restore WIP + surgical require bump (same as cascade pin path).
	saved, err := saveGoModSumSnap(modDir)
	if err != nil {
		return fmt.Errorf("wrk: dep-update save go.mod/go.sum in %s: %w", modDir, err)
	}
	usePartial := false
	dirty, err := goModSumUncommittedAt(checkout, modDir)
	if err != nil {
		return err
	}
	if dirty {
		usePartial = true
		if err := writeBaseGoModSum(checkout, modDir); err != nil {
			_ = restoreGoModSumSnap(modDir, saved)
			return fmt.Errorf("wrk: dep-update restore Base go.mod/go.sum in %s: %w", modDir, err)
		}
	}
	pinFail := func(err error) error {
		_ = restoreGoModSumSnap(modDir, saved)
		return err
	}
	if err := cascadePinKeepLocalReplace(modDir, a.DepModule, ver, dep, r.byLabel); err != nil {
		return pinFail(err)
	}
	if err := goModTidyForCascadePin(modDir, saved, usePartial, a.DepModule, depDir); err != nil {
		return pinFail(err)
	}
	_ = expandGoModRequireBlocks(filepath.Join(modDir, "go.mod"))
	pinSum := readGoSumFile(modDir)
	if err := cascadeCommitPin(checkout, modDir, a.DepModule, dep.LatestTag, ver, false); err != nil {
		return pinFail(err)
	}
	if usePartial {
		if err := restoreGoModSumSnap(modDir, saved); err != nil {
			return fmt.Errorf("wrk: dep-update restore WIP go.mod/go.sum in %s: %w", modDir, err)
		}
		if err := cascadePinKeepLocalReplace(modDir, a.DepModule, ver, dep, r.byLabel); err != nil {
			_ = restoreGoModSumSnap(modDir, saved)
			return fmt.Errorf("wrk: dep-update surgical pin in %s: %w", modDir, err)
		}
		_ = expandGoModRequireBlocks(filepath.Join(modDir, "go.mod"))
		if err := mergePinGoSumHashes(modDir, pinSum, a.DepModule, ver); err != nil {
			_ = restoreGoModSumSnap(modDir, saved)
			return fmt.Errorf("wrk: dep-update surgical go.sum in %s: %w", modDir, err)
		}
	}
	if r.stats != nil {
		r.stats.Pinned++
	}
	r.addMain(m.MainRepo)
	if m.MainRepo == "" {
		r.addMain(m.Path)
	}
	return nil
}

func (r *jobRunner) applyShip(a *Action, io HostIO) error {
	m, ok := r.byLabel[a.Lane]
	if !ok {
		return nil
	}
	main := m.MainRepo
	if main == "" {
		main = m.Path
	}
	main = storage.NormalizePath(main)
	switch a.Mode {
	case ModePush:
		tags := r.tagsForLane(a.Lane)
		if err := runPushMain(main, false, r.flags.Force, tags, io); err != nil {
			if isNoPushRemoteErr(err) {
				fmt.Fprintf(io.Err(), "warning: skip push %s: %v\n", main, err)
				return nil
			}
			return err
		}
		if r.stats != nil {
			r.stats.Pushed++
		}
	case ModeSync:
		_, err := runSyncWithColor(main, false, r.flags.Color, r.flags.NoColor, io)
		return err
	case ModeReinstall:
		n, err := runUnwindReinstallLocal(main, r.flags.Color, r.flags.NoColor, io)
		if err != nil {
			return err
		}
		if r.stats != nil {
			r.stats.Reinstalled += n
		}
	}
	return nil
}

func (r *jobRunner) addMain(path string) {
	path = storage.NormalizePath(path)
	if path == "" {
		return
	}
	if r.seenMain == nil {
		r.seenMain = map[string]struct{}{}
	}
	if _, ok := r.seenMain[path]; ok {
		return
	}
	r.seenMain[path] = struct{}{}
	r.shipMains = append(r.shipMains, path)
}
