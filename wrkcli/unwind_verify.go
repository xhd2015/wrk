package wrkcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UnwindVerifyReport is the read-only post-job audit for --unwind --verify.
type UnwindVerifyReport struct {
	WorkDir  string               `json:"work_dir"`
	Checks   []UnwindVerifyCheck  `json:"checks"`
	Summary  UnwindVerifySummary  `json:"summary"`
	Warnings []string             `json:"warnings"`
}

// UnwindVerifyCheck is one catalog audit line.
type UnwindVerifyCheck struct {
	ID       string   `json:"id"`
	Severity string   `json:"severity"`
	Status   string   `json:"status"` // pass | fail
	Count    *int     `json:"count,omitempty"`
	Details  []string `json:"details,omitempty"`
}

// UnwindVerifySummary aggregates check outcomes.
type UnwindVerifySummary struct {
	Checks int    `json:"checks"`
	Pass   int    `json:"pass"`
	Fail   int    `json:"fail"`
	Warn   int    `json:"warn"`
	Result string `json:"result"` // pass | fail
}

// verifyCheckCatalog is the locked error-severity check order for human/JSON.
var verifyCheckCatalog = []string{
	"dirty-peel",
	"needs-land",
	"owned-changed",
	"require-drift",
	"droppable-replace",
	"cascade-pending",
}

// BuildUnwindVerifyReport collects inventory, peel plan, module graph, cascade
// plan, and materializes the six error-severity checks. Does not mutate.
func BuildUnwindVerifyReport(workDir string) (*UnwindVerifyReport, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	inv, err := CollectStackInventory(cwd)
	if err != nil {
		return nil, err
	}
	members := inv.Members
	edges, err := BuildRepoDAG(members)
	if err != nil {
		return nil, err
	}
	edges = mergeRepoEdges(edges, inv.SyntheticEdges)
	plan, err := PlanUnwind(members, edges)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		plan = &UnwindPlan{}
	}

	byLabel := pickPeelMembersByLabel(members)
	modNodes, modEdges, err := buildUnwindModuleGraph(members, byLabel)
	if err != nil {
		return nil, err
	}
	attachTagScopeToModules(modNodes, members, nil)

	cascade, err := planUnwindCascadeFromGraph(modNodes, modEdges)
	if err != nil {
		return nil, err
	}
	if cascade == nil {
		cascade = &UnwindCascadePlan{}
	}

	warnings := append([]string(nil), inv.Warnings...)
	checks := materializeUnwindVerifyChecks(plan, modNodes, modEdges, cascade)

	pass, fail := 0, 0
	for _, c := range checks {
		switch c.Status {
		case "pass":
			pass++
		case "fail":
			fail++
		}
	}
	result := "pass"
	if fail > 0 {
		result = "fail"
	}

	return &UnwindVerifyReport{
		WorkDir: cwd,
		Checks:  checks,
		Summary: UnwindVerifySummary{
			Checks: len(checks),
			Pass:   pass,
			Fail:   fail,
			Warn:   len(warnings),
			Result: result,
		},
		Warnings: warnings,
	}, nil
}

func materializeUnwindVerifyChecks(
	plan *UnwindPlan,
	modNodes []UnwindGraphModuleNode,
	modEdges []UnwindGraphModuleEdge,
	cascade *UnwindCascadePlan,
) []UnwindVerifyCheck {
	// --- dirty-peel ---
	dirtyPeelFail := len(plan.PeelOrder) > 0
	dirtyPeel := UnwindVerifyCheck{
		ID:       "dirty-peel",
		Severity: "error",
		Status:   statusPassFail(dirtyPeelFail),
	}
	if dirtyPeelFail {
		dirtyPeel.Details = []string{
			fmt.Sprintf("%d dirty stack member(s) would still peel", len(plan.PeelOrder)),
		}
	}

	// --- needs-land ---
	needsLandFail := plan.NeedsLand
	needsLand := UnwindVerifyCheck{
		ID:       "needs-land",
		Severity: "error",
		Status:   statusPassFail(needsLandFail),
	}
	if needsLandFail {
		needsLand.Details = []string{"linked dirty stack still needs land (merge-back)"}
	}

	// --- owned-changed ---
	ownedCount := 0
	var ownedDetails []string
	for _, n := range modNodes {
		if n.OwnedChanged || n.NextTag != "" {
			ownedCount++
			if n.Path != "" {
				msg := n.Path
				if n.NextTag != "" {
					msg += " next=" + tagVersion(n.NextTag)
				}
				ownedDetails = append(ownedDetails, msg)
			}
		}
	}
	ownedFail := ownedCount > 0
	ownedChanged := UnwindVerifyCheck{
		ID:       "owned-changed",
		Severity: "error",
		Status:   statusPassFail(ownedFail),
		Count:    intPtr(ownedCount),
	}
	if ownedFail {
		ownedChanged.Details = ownedDetails
	}

	// --- require-drift ---
	nodeByPath := make(map[string]UnwindGraphModuleNode, len(modNodes))
	for _, n := range modNodes {
		if n.Path != "" {
			nodeByPath[n.Path] = n
		}
	}
	driftCount := 0
	var driftDetails []string
	for _, e := range modEdges {
		if e.Kind != "require" || e.From == "" || e.To == "" {
			continue
		}
		dep, ok := nodeByPath[e.To]
		if !ok {
			continue
		}
		if dep.LatestTag == "" || e.Version == "" {
			continue
		}
		if versionsMatch(e.Version, dep.LatestTag) {
			continue
		}
		driftCount++
		lv := tagVersion(dep.LatestTag)
		if lv == "" {
			lv = dep.LatestTag
		}
		driftDetails = append(driftDetails,
			fmt.Sprintf("%s → %s require %s (latest %s)", e.From, e.To, e.Version, lv))
	}
	driftFail := driftCount > 0
	requireDrift := UnwindVerifyCheck{
		ID:       "require-drift",
		Severity: "error",
		Status:   statusPassFail(driftFail),
		Count:    intPtr(driftCount),
	}
	if driftFail {
		requireDrift.Details = driftDetails
	}

	// --- droppable-replace ---
	dropCount := 0
	var dropDetails []string
	for _, e := range modEdges {
		if e.Kind != "replace" {
			continue
		}
		from, ok1 := nodeByPath[e.From]
		to, ok2 := nodeByPath[e.To]
		if !ok1 || !ok2 {
			continue
		}
		if !isDroppableExternalStackReplace(from, to, e) {
			continue
		}
		dropCount++
		dropDetails = append(dropDetails,
			fmt.Sprintf("%s → %s droppable external replace", e.From, e.To))
	}
	dropFail := dropCount > 0
	droppableReplace := UnwindVerifyCheck{
		ID:       "droppable-replace",
		Severity: "error",
		Status:   statusPassFail(dropFail),
		Count:    intPtr(dropCount),
	}
	if dropFail {
		droppableReplace.Details = dropDetails
	}

	// --- cascade-pending ---
	cascadeSteps := 0
	if cascade != nil {
		cascadeSteps = len(cascade.Steps)
	}
	cascadeFail := cascadeSteps > 0
	cascadePending := UnwindVerifyCheck{
		ID:       "cascade-pending",
		Severity: "error",
		Status:   statusPassFail(cascadeFail),
	}
	if cascadeFail {
		cascadePending.Details = []string{
			fmt.Sprintf("%d cascade step(s) still planned (tag-next/pin)", cascadeSteps),
		}
	}

	// Fixed catalog order.
	byID := map[string]UnwindVerifyCheck{
		"dirty-peel":        dirtyPeel,
		"needs-land":        needsLand,
		"owned-changed":     ownedChanged,
		"require-drift":     requireDrift,
		"droppable-replace": droppableReplace,
		"cascade-pending":   cascadePending,
	}
	out := make([]UnwindVerifyCheck, 0, len(verifyCheckCatalog))
	for _, id := range verifyCheckCatalog {
		out = append(out, byID[id])
	}
	return out
}

func statusPassFail(fail bool) string {
	if fail {
		return "fail"
	}
	return "pass"
}

func intPtr(n int) *int {
	return &n
}

// FormatUnwindVerifyHuman renders the human verify body (trailing newline).
// colorOn: green pass/result pass; red FAIL/result fail; gray banners.
func FormatUnwindVerifyHuman(report *UnwindVerifyReport, colorOn bool) string {
	if report == nil {
		return ""
	}
	var b strings.Builder

	banner := func(s string) string {
		return paint(s, ansiGrey, colorOn)
	}
	statusTok := func(status string) string {
		switch status {
		case "pass":
			return paint("pass", ansiGreen, colorOn)
		case "fail":
			return paint("FAIL", ansiRed, colorOn)
		default:
			return status
		}
	}

	b.WriteString(banner("==== unwind verify ===="))
	b.WriteByte('\n')

	// Fixed id column width for scanability (longest catalog id = droppable-replace).
	const idWidth = 19
	for _, c := range report.Checks {
		id := padRightVisible(c.ID, idWidth)
		line := "  " + id + "  " + statusTok(c.Status)
		// Optional count annotations for selected checks.
		if c.Count != nil {
			switch c.ID {
			case "owned-changed":
				line += fmt.Sprintf("   (%d modules)", *c.Count)
			case "require-drift":
				line += fmt.Sprintf("   (%d edges)", *c.Count)
			case "droppable-replace":
				line += fmt.Sprintf("   (%d replaces)", *c.Count)
			}
		}
		b.WriteString(line)
		b.WriteByte('\n')
		if c.Status == "fail" {
			for _, d := range c.Details {
				fmt.Fprintf(&b, "    %s\n", d)
			}
		}
	}
	b.WriteByte('\n')

	b.WriteString(banner("==== status summary ===="))
	b.WriteByte('\n')
	s := report.Summary
	fmt.Fprintf(&b, "checks: %d  pass: %d  fail: %d  warn: %d\n",
		s.Checks, s.Pass, s.Fail, s.Warn)
	resultLine := "result: "
	switch s.Result {
	case "pass":
		resultLine += paint("pass", ansiGreen, colorOn)
	case "fail":
		resultLine += paint("fail", ansiRed, colorOn)
	default:
		resultLine += s.Result
	}
	b.WriteString(resultLine)
	b.WriteByte('\n')
	return b.String()
}

// FormatUnwindVerifyJSON renders pure JSON (trailing newline, no ANSI).
func FormatUnwindVerifyJSON(report *UnwindVerifyReport) ([]byte, error) {
	if report == nil {
		return []byte("{}\n"), nil
	}
	checks := report.Checks
	if checks == nil {
		checks = []UnwindVerifyCheck{}
	}
	warnings := report.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	// Encode without Details for cleaner JSON (optional field omitted when empty).
	type checkJSON struct {
		ID       string `json:"id"`
		Severity string `json:"severity"`
		Status   string `json:"status"`
		Count    *int   `json:"count,omitempty"`
	}
	cj := make([]checkJSON, 0, len(checks))
	for _, c := range checks {
		cj = append(cj, checkJSON{
			ID:       c.ID,
			Severity: c.Severity,
			Status:   c.Status,
			Count:    c.Count,
		})
	}
	out := struct {
		WorkDir  string              `json:"work_dir"`
		Checks   []checkJSON         `json:"checks"`
		Summary  UnwindVerifySummary `json:"summary"`
		Warnings []string            `json:"warnings"`
	}{
		WorkDir:  report.WorkDir,
		Checks:   cj,
		Summary:  report.Summary,
		Warnings: warnings,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}

// runUnwindVerify is the read-only verify path: report only, no apply.
// Logical FAIL: full report on stdout, ExitCodeError(1), no Error: prefix.
// Fatal preflight errors propagate as normal errors (Error: on stderr).
func runUnwindVerify(workDir string, jsonOut bool, colorOn bool) error {
	report, err := BuildUnwindVerifyReport(workDir)
	if err != nil {
		return err
	}
	// Soft inventory warnings on stderr (same prefix policy as show-graph).
	for _, w := range report.Warnings {
		msg := w
		if !strings.HasPrefix(msg, "warning:") && !strings.HasPrefix(msg, "Warning:") {
			msg = "warning: " + msg
		}
		fmt.Fprintln(os.Stderr, msg)
	}
	if jsonOut {
		data, err := FormatUnwindVerifyJSON(report)
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprint(os.Stdout, FormatUnwindVerifyHuman(report, colorOn)); err != nil {
			return err
		}
	}
	if report.Summary.Result == "fail" {
		return ExitCodeError{Code: 1}
	}
	return nil
}
