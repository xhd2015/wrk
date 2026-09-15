package wrkcli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/installplan"
	"golang.org/x/term"
)

// Method and Action are string type aliases so harness code can use string(it.Method).
type Method = installplan.Method
type Action string

const (
	MethodGoInstall           = installplan.MethodGoInstall
	MethodGoRunInstall        = installplan.MethodGoRunInstall
	ActionInstall      Action = "install"
	ActionSkip         Action = "skip"

	DiagLevelNotice  = installplan.DiagLevelNotice
	DiagLevelWarning = installplan.DiagLevelWarning

	DiagKindPreferScript    = installplan.DiagKindPreferScript
	DiagKindAmbiguousCmd    = installplan.DiagKindAmbiguousCmd
	DiagKindAmbiguousScript = installplan.DiagKindAmbiguousScript
	DiagKindNestedScript    = installplan.DiagKindNestedScript
)

// ReinstallDiagnostic is a non-fatal notice or warning produced while merging
// cmd/ and script/ candidates for the same BinName.
type ReinstallDiagnostic = installplan.Diagnostic

// LocalReinstallPlan is the pure discovery/filter result for local binary reinstalls.
type LocalReinstallPlan struct {
	ModuleRoot  string
	ModulePath  string // full module path from go.mod
	ModuleName  string // basename of module path from go.mod
	BinDir      string
	Items       []PlanItem            // sorted lexicographically by BinName
	Diagnostics []ReinstallDiagnostic // sorted by BinName, then Kind
}

// MultiLocalReinstallPlan is the multi-module discovery/filter result for a
// shared binDir. Modules are ordered lexicographically by absolute ModuleRoot.
type MultiLocalReinstallPlan struct {
	BinDir  string
	Modules []ModuleReinstallPlan
}

// ModuleReinstallPlan is one module's contribution to a multi-module plan.
type ModuleReinstallPlan struct {
	ModuleRoot  string
	ModulePath  string                // full module path from go.mod
	ModuleName  string                // basename of module path from go.mod
	RelDir      string                // module root relative to scan root ("." or slash path); set by FromWorkDir
	Items       []PlanItem            // sorted lexicographically by BinName
	Diagnostics []ReinstallDiagnostic // per-module; sorted by BinName, then Kind
}

// PlanItem is one candidate binary to install or skip.
type PlanItem struct {
	BinName string
	RelPath string // "./cmd/foo" or "./script/foo/install"
	Method  Method // "go-install" | "go-run-install"
	Action  Action // "install" | "skip"
}

// PlanLocalReinstalls discovers package-main candidates under moduleRoot's
// cmd/ and script/ trees and filters them against binDir.
//
// moduleRoot must contain a parseable go.mod with a module path.
// Callers resolve binDir (e.g. GOBIN); this function only stats entries there.
// Discovery and merge rules live in gotool/mod/installplan.
func PlanLocalReinstalls(moduleRoot, binDir string) (*LocalReinstallPlan, error) {
	ip, err := installplan.Discover(moduleRoot)
	if err != nil {
		return nil, err
	}
	mod := modulePlanFromInstall(*ip, binDir)
	return &LocalReinstallPlan{
		ModuleRoot:  mod.ModuleRoot,
		ModulePath:  mod.ModulePath,
		ModuleName:  mod.ModuleName,
		BinDir:      binDir,
		Items:       mod.Items,
		Diagnostics: mod.Diagnostics,
	}, nil
}

func modulePlanFromInstall(ip installplan.ModulePlan, binDir string) ModuleReinstallPlan {
	items := make([]PlanItem, 0, len(ip.Items))
	for _, it := range ip.Items {
		items = append(items, PlanItem{
			BinName: it.BinName,
			RelPath: it.RelPath,
			Method:  it.Method,
			Action:  binAction(binDir, it.BinName),
		})
	}
	return ModuleReinstallPlan{
		ModuleRoot:  ip.ModuleRoot,
		ModulePath:  ip.ModulePath,
		ModuleName:  ip.ModuleName,
		RelDir:      ip.RelDir,
		Items:       items,
		Diagnostics: ip.Diagnostics,
	}
}

// PlanLocalReinstallsMulti runs PlanLocalReinstalls for each module root and
// returns a multi-module plan sorted by absolute ModuleRoot path.
//
// Empty moduleRoots yields an empty Modules list and nil error.
// After per-module planning, if the same BinName has Action=install from two
// or more modules, returns a hard error naming the bin and both modules.
// Skip-only (or install×skip) duplicates across modules are allowed.
func PlanLocalReinstallsMulti(moduleRoots []string, binDir string) (*MultiLocalReinstallPlan, error) {
	ip, err := installplan.DiscoverMulti(moduleRoots)
	if err != nil {
		return nil, err
	}
	modules := make([]ModuleReinstallPlan, 0, len(ip.Modules))
	for _, m := range ip.Modules {
		modules = append(modules, modulePlanFromInstall(m, binDir))
	}
	if err := detectCrossModuleInstallCollisions(modules); err != nil {
		return nil, err
	}
	return &MultiLocalReinstallPlan{
		BinDir:  binDir,
		Modules: modules,
	}, nil
}

// ResolveReinstallScanRoot returns the absolute directory from which Go module
// discovery should begin for multi-module local reinstall planning.
func ResolveReinstallScanRoot(workDir string, useMain bool) (string, error) {
	return installplan.ResolveScanRoot(workDir, useMain)
}

// PlanLocalReinstallsFromWorkDir resolves the scan root from workDir, discovers
// every Go module under that root via mod/scan, and builds a multi-module
// reinstall plan against binDir.
func PlanLocalReinstallsFromWorkDir(workDir, binDir string, useMain bool) (*MultiLocalReinstallPlan, error) {
	ip, err := installplan.DiscoverFromWorkDir(workDir, useMain)
	if err != nil {
		return nil, err
	}
	modules := make([]ModuleReinstallPlan, 0, len(ip.Modules))
	for _, m := range ip.Modules {
		modules = append(modules, modulePlanFromInstall(m, binDir))
	}
	if err := detectCrossModuleInstallCollisions(modules); err != nil {
		return nil, err
	}
	return &MultiLocalReinstallPlan{
		BinDir:  binDir,
		Modules: modules,
	}, nil
}

// installClaim tracks one module that wants to install a given bin.
type installClaim struct {
	ModuleRoot string
	ModuleName string
}

// detectCrossModuleInstallCollisions returns an error when the same BinName has
// Action=install in two or more modules.
func detectCrossModuleInstallCollisions(modules []ModuleReinstallPlan) error {
	claims := make(map[string][]installClaim)
	for _, m := range modules {
		for _, it := range m.Items {
			if it.Action != ActionInstall {
				continue
			}
			claims[it.BinName] = append(claims[it.BinName], installClaim{
				ModuleRoot: m.ModuleRoot,
				ModuleName: m.ModuleName,
			})
		}
	}
	var collidingBins []string
	for bin, list := range claims {
		if len(list) >= 2 {
			collidingBins = append(collidingBins, bin)
		}
	}
	if len(collidingBins) == 0 {
		return nil
	}
	sort.Strings(collidingBins)
	bin := collidingBins[0]
	list := claims[bin]
	a, b := list[0], list[1]
	return fmt.Errorf(
		"bin %q claimed for install by multiple modules: %s (%s) and %s (%s)",
		bin, a.ModuleRoot, a.ModuleName, b.ModuleRoot, b.ModuleName,
	)
}

// binAction returns install if $binDir/binName is a regular file or a symlink
// that resolves to a file; otherwise skip.
func binAction(binDir, binName string) Action {
	path := filepath.Join(binDir, binName)
	// Use Lstat first? Spec: "exists as a regular file or a symlink that resolves to a file"
	// os.Stat follows symlinks.
	info, err := os.Stat(path)
	if err != nil {
		return ActionSkip
	}
	if info.IsDir() {
		return ActionSkip
	}
	// Mode().IsRegular() is false for some special files; after Stat of symlink
	// target, a normal file is regular. Accept any non-dir as file-like.
	if info.Mode().IsRegular() {
		return ActionInstall
	}
	// Fallback: if it's not a dir and Stat succeeded, treat as present file
	// (e.g. some platforms). Prefer regular-only for safety.
	return ActionSkip
}

// ReinstallExecStats counts execute-path outcomes for --reinstall-local.
type ReinstallExecStats struct {
	Reinstalled int
	Skipped     int
	Failed      int
}

// runReinstallLocal implements wrk --reinstall-local [--dry-run] [--main] [--color].
// dry-run prints the plan and does not run go install/run.
// Without --dry-run, installs run sequentially (continue on failure).
//
// Planning uses PlanLocalReinstallsFromWorkDir(workDir, binDir, useMain).
// useMain=false scans the worktree toplevel (or walk-up); useMain=true
// (from --main) scans the main repository of this checkout.
// colorFlag forces ANSI on diagnostic prefixes when true.
// names, when non-empty, selects only those bins and forces install (no binDir gate).
func runReinstallLocal(workDir string, dryRun bool, useMain bool, colorFlag bool, names []string) error {
	_, err := runReinstallLocalEx(workDir, dryRun, useMain, colorFlag, false, names)
	return err
}

// runReinstallLocalEx is runReinstallLocal with NoColor and returned execute stats.
func runReinstallLocalEx(workDir string, dryRun bool, useMain bool, colorFlag, noColor bool, names []string) (ReinstallExecStats, error) {
	return runReinstallLocalExTo(workDir, dryRun, useMain, colorFlag, noColor, names, os.Stdout, os.Stderr)
}

// runReinstallLocalExTo is runReinstallLocalEx with custom writers (unwind HostIO).
func runReinstallLocalExTo(workDir string, dryRun bool, useMain bool, colorFlag, noColor bool, names []string, out, errW io.Writer) (ReinstallExecStats, error) {
	var empty ReinstallExecStats
	if out == nil {
		out = os.Stdout
	}
	if errW == nil {
		errW = os.Stderr
	}
	binDir, err := resolveLocalReinstallBinDir()
	if err != nil {
		return empty, err
	}
	plan, err := PlanLocalReinstallsFromWorkDir(workDir, binDir, useMain)
	if err != nil {
		return empty, err
	}
	if len(names) > 0 {
		plan, err = filterNamedReinstallPlan(plan, names)
		if err != nil {
			return empty, err
		}
	}
	diagColor := reinstallDiagColorEnabled(colorFlag)
	stdoutColor := resolveStdoutColor(colorFlag, noColor)
	if dryRun {
		return empty, printMultiLocalReinstallDryRun(plan, diagColor)
	}
	return executeMultiLocalReinstallsTo(plan, diagColor, stdoutColor, out, errW)
}

// filterNamedReinstallPlan keeps only the requested bin names from a multi plan.
// Each surviving item is forced to ActionInstall (named mode skips the binDir gate).
// Missing names, empty names, and the same name claimed by multiple modules are errors.
func filterNamedReinstallPlan(plan *MultiLocalReinstallPlan, names []string) (*MultiLocalReinstallPlan, error) {
	if plan == nil {
		return nil, fmt.Errorf("nil reinstall plan")
	}
	ordered, err := dedupeReinstallNames(names)
	if err != nil {
		return nil, err
	}

	type hit struct {
		modIdx  int
		itemIdx int
	}
	byName := make(map[string][]hit)
	for mi, mod := range plan.Modules {
		for ii, it := range mod.Items {
			byName[it.BinName] = append(byName[it.BinName], hit{modIdx: mi, itemIdx: ii})
		}
	}

	// modIdx -> selected items (forced install), preserving request order within module.
	selected := make([][]PlanItem, len(plan.Modules))
	diagKeep := make([]map[string]struct{}, len(plan.Modules))
	for i := range diagKeep {
		diagKeep[i] = make(map[string]struct{})
	}

	for _, name := range ordered {
		hits := byName[name]
		if len(hits) == 0 {
			// Prefer ambiguous diagnostic wording when the bin was omitted for that reason.
			if paths, ok := namedAmbiguousPaths(plan, name); ok {
				return nil, fmt.Errorf("wrk: --reinstall-local: bin %q is ambiguous (%s)", name, strings.Join(paths, ", "))
			}
			return nil, fmt.Errorf("wrk: --reinstall-local: no install candidate for %q", name)
		}
		if len(hits) > 1 {
			a := plan.Modules[hits[0].modIdx]
			b := plan.Modules[hits[1].modIdx]
			return nil, fmt.Errorf(
				"wrk: --reinstall-local: bin %q claimed by multiple modules: %s (%s) and %s (%s)",
				name, a.ModuleRoot, a.ModuleName, b.ModuleRoot, b.ModuleName,
			)
		}
		h := hits[0]
		it := plan.Modules[h.modIdx].Items[h.itemIdx]
		it.Action = ActionInstall
		selected[h.modIdx] = append(selected[h.modIdx], it)
		diagKeep[h.modIdx][name] = struct{}{}
	}

	outMods := make([]ModuleReinstallPlan, 0, len(plan.Modules))
	for mi, mod := range plan.Modules {
		items := selected[mi]
		if len(items) == 0 {
			continue
		}
		var diags []ReinstallDiagnostic
		for _, d := range mod.Diagnostics {
			if _, ok := diagKeep[mi][d.BinName]; ok {
				diags = append(diags, d)
			}
		}
		outMods = append(outMods, ModuleReinstallPlan{
			ModuleRoot:  mod.ModuleRoot,
			ModulePath:  mod.ModulePath,
			ModuleName:  mod.ModuleName,
			RelDir:      mod.RelDir,
			Items:       items,
			Diagnostics: diags,
		})
	}
	return &MultiLocalReinstallPlan{
		BinDir:  plan.BinDir,
		Modules: outMods,
	}, nil
}

func dedupeReinstallNames(names []string) ([]string, error) {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		if n == "" {
			return nil, fmt.Errorf("wrk: --reinstall-local: empty name")
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("wrk: --reinstall-local: empty name list")
	}
	return out, nil
}

// namedAmbiguousPaths returns sorted paths from ambiguous-* diagnostics for bin, if any.
func namedAmbiguousPaths(plan *MultiLocalReinstallPlan, bin string) ([]string, bool) {
	var paths []string
	seen := make(map[string]struct{})
	for _, mod := range plan.Modules {
		for _, d := range mod.Diagnostics {
			if d.BinName != bin {
				continue
			}
			if d.Kind != DiagKindAmbiguousCmd && d.Kind != DiagKindAmbiguousScript {
				continue
			}
			for _, p := range d.Paths {
				if _, ok := seen[p]; ok {
					continue
				}
				seen[p] = struct{}{}
				paths = append(paths, p)
			}
		}
	}
	if len(paths) == 0 {
		return nil, false
	}
	sort.Strings(paths)
	return paths, true
}

// reinstallDiagColorEnabled reports whether diagnostic prefix tokens should use ANSI.
// --color forces on; otherwise on only when stderr is a TTY and NO_COLOR is empty.
func reinstallDiagColorEnabled(colorFlag bool) bool {
	if colorFlag {
		return true
	}
	return term.IsTerminal(int(os.Stderr.Fd())) && os.Getenv("NO_COLOR") == ""
}

// findModuleRootWalking walks up from start looking for a go.mod file.
func findModuleRootWalking(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found walking up from %s", start)
		}
		dir = parent
	}
}

// resolveLocalReinstallBinDir returns GOBIN if set, else $(go env GOPATH)/bin.
func resolveLocalReinstallBinDir() (string, error) {
	if gobin := strings.TrimSpace(os.Getenv("GOBIN")); gobin != "" {
		abs, err := filepath.Abs(gobin)
		if err != nil {
			return "", fmt.Errorf("resolve GOBIN: %w", err)
		}
		return abs, nil
	}
	gopath, err := goEnvGOPATH()
	if err != nil {
		return "", err
	}
	if gopath == "" {
		return "", fmt.Errorf("go env GOPATH returned empty")
	}
	// GOPATH may be a list; the first entry is the default install target.
	first := gopath
	if i := strings.Index(gopath, string(os.PathListSeparator)); i >= 0 {
		first = gopath[:i]
	}
	abs, err := filepath.Abs(filepath.Join(first, "bin"))
	if err != nil {
		return "", fmt.Errorf("resolve GOPATH/bin: %w", err)
	}
	return abs, nil
}

func goEnvGOPATH() (string, error) {
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", fmt.Errorf("go env GOPATH: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// printLocalReinstallDryRun writes would:/skip: lines and a summary to stdout
// for a single-module plan (legacy helper; CLI uses printMultiLocalReinstallDryRun).
func printLocalReinstallDryRun(plan *LocalReinstallPlan, colorOn bool) error {
	printReinstallDiagnostics(plan.Diagnostics, colorOn)
	nInstall, nSkip := 0, 0
	if err := printPlanItemsDryRun(plan.ModuleRoot, plan.Items, plan.BinDir, &nInstall, &nSkip); err != nil {
		return err
	}
	fmt.Printf("would: reinstall %d binaries (%d skipped)\n", nInstall, nSkip)
	return nil
}

// printMultiLocalReinstallDryRun prints a multi-module dry-run plan.
//
// Diagnostics (if any) are printed to stderr first, then plan lines on stdout.
// K==1: same format as single-module dry-run (no # module headers; summary
// without "across").
// K>1: for each module in plan order, "# module <ModulePath> (<RelDir>)" then
// that module's would:/skip: lines; summary ends with " across K modules".
func printMultiLocalReinstallDryRun(plan *MultiLocalReinstallPlan, colorOn bool) error {
	k := len(plan.Modules)
	// Print all diagnostics first (module order, already sorted within each).
	for _, mod := range plan.Modules {
		printReinstallDiagnostics(mod.Diagnostics, colorOn)
	}
	nInstall, nSkip := 0, 0
	for _, mod := range plan.Modules {
		if k > 1 {
			modulePath := mod.ModulePath
			if modulePath == "" {
				modulePath = mod.ModuleName
			}
			relDir := mod.RelDir
			if relDir == "" {
				relDir = "."
			}
			fmt.Printf("# module %s (%s)\n", modulePath, relDir)
		}
		if err := printPlanItemsDryRun(mod.ModuleRoot, mod.Items, plan.BinDir, &nInstall, &nSkip); err != nil {
			return err
		}
	}
	if k > 1 {
		fmt.Printf("would: reinstall %d binaries (%d skipped) across %d modules\n", nInstall, nSkip, k)
	} else {
		fmt.Printf("would: reinstall %d binaries (%d skipped)\n", nInstall, nSkip)
	}
	return nil
}

// printReinstallDiagnostics writes one stderr line per diagnostic to os.Stderr.
// Color (when colorOn) applies only to the "notice:" / "warning:" prefix.
func printReinstallDiagnostics(diags []ReinstallDiagnostic, colorOn bool) {
	printReinstallDiagnosticsTo(os.Stderr, diags, colorOn)
}

// printReinstallDiagnosticsTo writes one diagnostic line per entry to errW
// (nil → os.Stderr). Used by concurrent ship so diagnostics hit the lane sink
// instead of the real terminal while the TTY spinner owns the screen.
func printReinstallDiagnosticsTo(errW io.Writer, diags []ReinstallDiagnostic, colorOn bool) {
	if errW == nil {
		errW = os.Stderr
	}
	for _, d := range diags {
		fmt.Fprint(errW, formatReinstallDiagnosticLine(d, colorOn))
	}
}

// formatReinstallDiagnosticLine returns the full stderr line including trailing \n.
func formatReinstallDiagnosticLine(d ReinstallDiagnostic, colorOn bool) string {
	var prefix string
	var body string
	switch d.Kind {
	case DiagKindPreferScript:
		prefix = "notice:"
		winner, rest := splitWinner(d.Paths)
		body = fmt.Sprintf(" bin %s: preferring %s over %s", d.BinName, winner, strings.Join(rest, ", "))
	case DiagKindNestedScript:
		prefix = "warning:"
		winner, rest := splitWinner(d.Paths)
		body = fmt.Sprintf(" bin %s: ignoring nested script (%s); using %s", d.BinName, strings.Join(rest, ", "), winner)
	case DiagKindAmbiguousCmd:
		prefix = "warning:"
		body = fmt.Sprintf(" bin %s: ambiguous under cmd (%s); skipping", d.BinName, strings.Join(d.Paths, ", "))
	case DiagKindAmbiguousScript:
		prefix = "warning:"
		body = fmt.Sprintf(" bin %s: ambiguous under script (%s); skipping", d.BinName, strings.Join(d.Paths, ", "))
	default:
		prefix = "warning:"
		body = fmt.Sprintf(" bin %s: %s", d.BinName, d.Kind)
	}
	if colorOn {
		switch d.Level {
		case DiagLevelNotice:
			prefix = colorize(prefix, ansiGrey)
		case DiagLevelWarning:
			prefix = colorize(prefix, ansiOrange)
		default:
			// Fall back by kind tokens already set.
			if d.Kind == DiagKindPreferScript {
				prefix = colorize("notice:", ansiGrey)
			} else {
				prefix = colorize(prefix, ansiOrange)
			}
		}
	}
	return prefix + body + "\n"
}

func splitWinner(paths []string) (winner string, rest []string) {
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], paths[1:]
}

// printPlanItemsDryRun writes would:/skip: lines for items; accumulates counters.
// Install paths are re-rooted to the nearest go.mod under moduleRoot so dry-run
// matches what execute will run (e.g. ./foo when cmd/ is a nested module).
func printPlanItemsDryRun(moduleRoot string, items []PlanItem, binDir string, nInstall, nSkip *int) error {
	for _, it := range items {
		switch it.Action {
		case ActionInstall:
			*nInstall++
			_, ownRel, err := resolveGoPackageRoot(moduleRoot, it.RelPath)
			if err != nil {
				return err
			}
			switch it.Method {
			case MethodGoInstall:
				fmt.Printf("would: go install %s\n", ownRel)
			case MethodGoRunInstall:
				fmt.Printf("would: go run %s\n", ownRel)
			default:
				return fmt.Errorf("unknown reinstall method %q for %s", it.Method, it.BinName)
			}
		case ActionSkip:
			*nSkip++
			fmt.Printf("skip: %s (not in %s)\n", it.BinName, binDir)
		default:
			return fmt.Errorf("unknown reinstall action %q for %s", it.Action, it.BinName)
		}
	}
	return nil
}

// executeLocalReinstalls runs planned go install / go run commands sequentially
// for a single-module plan (legacy helper; CLI uses executeMultiLocalReinstalls).
// Skip items print the same skip: line as dry-run and do not invoke go.
// Child stdout/stderr are streamed to the process. Continues after failures.
// Summary: reinstalled N, skipped M, failed F. Soft: failed > 0 → stderr warning, exit 0.
func executeLocalReinstalls(plan *LocalReinstallPlan, diagColor, stdoutColor bool) error {
	printReinstallDiagnostics(plan.Diagnostics, diagColor)
	nReinstalled, nSkip, nFailed := 0, 0, 0
	if err := executePlanItems(plan.ModuleRoot, plan.BinDir, plan.Items, stdoutColor, &nReinstalled, &nSkip, &nFailed); err != nil {
		return err
	}
	fmt.Println(formatReinstallSummaryLine(nReinstalled, nSkip, nFailed, stdoutColor))
	if nFailed > 0 {
		printReinstallFailedWarning(nFailed, diagColor)
	}
	return nil
}

// executeMultiLocalReinstalls runs installs for every module in the multi plan
// (module order, then BinName order within each module). Same progress/skip
// lines and summary as single-module execute. Continues after failures.
// Soft: failed > 0 → stderr warning, exit 0 (hard plan errors still fail).
func executeMultiLocalReinstalls(plan *MultiLocalReinstallPlan, diagColor, stdoutColor bool) (ReinstallExecStats, error) {
	return executeMultiLocalReinstallsTo(plan, diagColor, stdoutColor, os.Stdout, os.Stderr)
}

func executeMultiLocalReinstallsTo(plan *MultiLocalReinstallPlan, diagColor, stdoutColor bool, out, errW io.Writer) (ReinstallExecStats, error) {
	if out == nil {
		out = os.Stdout
	}
	if errW == nil {
		errW = os.Stderr
	}
	var st ReinstallExecStats
	for _, mod := range plan.Modules {
		printReinstallDiagnosticsTo(errW, mod.Diagnostics, diagColor)
	}
	nReinstalled, nSkip, nFailed := 0, 0, 0
	for _, mod := range plan.Modules {
		if err := executePlanItemsTo(mod.ModuleRoot, plan.BinDir, mod.Items, stdoutColor, out, errW, &nReinstalled, &nSkip, &nFailed); err != nil {
			return st, err
		}
	}
	fmt.Fprintln(out, formatReinstallSummaryLine(nReinstalled, nSkip, nFailed, stdoutColor))
	if nFailed > 0 {
		printReinstallFailedWarningTo(errW, nFailed, diagColor)
	}
	st = ReinstallExecStats{Reinstalled: nReinstalled, Skipped: nSkip, Failed: nFailed}
	return st, nil
}

// printReinstallFailedWarning emits a soft-failure notice when installs fail.
// Install failures do not fail the process (exit 0); hard plan errors still do.
func printReinstallFailedWarning(nFailed int, colorOn bool) {
	printReinstallFailedWarningTo(os.Stderr, nFailed, colorOn)
}

func printReinstallFailedWarningTo(errW io.Writer, nFailed int, colorOn bool) {
	if errW == nil {
		errW = os.Stderr
	}
	prefix := "warning:"
	if colorOn {
		prefix = colorize(prefix, ansiOrange)
	}
	fmt.Fprintf(errW, "%s reinstall finished with %d failed\n", prefix, nFailed)
}

// executePlanItems runs install/skip actions for one module's items.
// Unknown method/action returns a hard error (stops the plan).
// Install/run paths are re-rooted to the nearest go.mod under moduleRoot
// (progress lines and cmd.Dir use the post-re-root path/root).
// stdoutColor greens the go install/run verb when true.
func executePlanItems(moduleRoot, binDir string, items []PlanItem, stdoutColor bool, nReinstalled, nSkip, nFailed *int) error {
	return executePlanItemsTo(moduleRoot, binDir, items, stdoutColor, os.Stdout, os.Stderr, nReinstalled, nSkip, nFailed)
}

func executePlanItemsTo(moduleRoot, binDir string, items []PlanItem, stdoutColor bool, out, errW io.Writer, nReinstalled, nSkip, nFailed *int) error {
	if out == nil {
		out = os.Stdout
	}
	if errW == nil {
		errW = os.Stderr
	}
	for _, it := range items {
		switch it.Action {
		case ActionInstall:
			ownRoot, ownRel, err := resolveGoPackageRoot(moduleRoot, it.RelPath)
			if err != nil {
				// Soft-fail this item so other installs can continue.
				*nFailed++
				continue
			}
			switch it.Method {
			case MethodGoInstall:
				fmt.Fprintln(out, formatGoInstallProgressLine(MethodGoInstall, ownRel, stdoutColor))
				err = runGoInModuleTo(ownRoot, "install", ownRel, out, errW)
			case MethodGoRunInstall:
				fmt.Fprintln(out, formatGoInstallProgressLine(MethodGoRunInstall, ownRel, stdoutColor))
				err = runGoInModuleTo(ownRoot, "run", ownRel, out, errW)
			default:
				return fmt.Errorf("unknown reinstall method %q for %s", it.Method, it.BinName)
			}
			if err != nil {
				*nFailed++
			} else {
				*nReinstalled++
			}
		case ActionSkip:
			*nSkip++
			fmt.Fprintf(out, "skip: %s (not in %s)\n", it.BinName, binDir)
		default:
			return fmt.Errorf("unknown reinstall action %q for %s", it.Action, it.BinName)
		}
	}
	return nil
}

// resolveGoPackageRoot returns the module directory owning pkg and the path
// relative to that module for go install/run.
//
// It walks up from the package directory looking for the nearest go.mod, without
// leaving moduleRoot. This re-roots installs when discovery planned under a
// parent module but the package lives in a nested module (e.g. cmd/go.mod +
// cmd/foo → ownRoot=cmd, ownRel=./foo).
func resolveGoPackageRoot(moduleRoot, relPath string) (ownRoot, ownRel string, err error) {
	moduleRoot = filepath.Clean(moduleRoot)
	if abs, aerr := filepath.Abs(moduleRoot); aerr == nil {
		moduleRoot = abs
	}
	if resolved, rerr := filepath.EvalSymlinks(moduleRoot); rerr == nil {
		moduleRoot = resolved
	}

	rel := strings.TrimPrefix(relPath, "./")
	pkgDir := filepath.Clean(filepath.Join(moduleRoot, filepath.FromSlash(rel)))
	if resolved, rerr := filepath.EvalSymlinks(pkgDir); rerr == nil {
		pkgDir = resolved
	}

	dir := pkgDir
	for {
		if !pathIsUnderOrEqual(dir, moduleRoot) {
			break
		}
		if _, serr := os.Stat(filepath.Join(dir, "go.mod")); serr == nil {
			ownRoot = dir
			relToPkg, rerr := filepath.Rel(ownRoot, pkgDir)
			if rerr != nil {
				return "", "", rerr
			}
			relToPkg = filepath.ToSlash(relToPkg)
			if relToPkg == "" || relToPkg == "." {
				ownRel = "."
			} else {
				ownRel = "./" + relToPkg
			}
			return ownRoot, ownRel, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("no go.mod found for package %s under module root %s", relPath, moduleRoot)
}

// pathIsUnderOrEqual reports whether path is the same as root or a descendant.
func pathIsUnderOrEqual(path, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// runGoInModule runs `go <subcmd> <relPath>` with Dir=moduleRoot and inherited env
// (including GOBIN). Streams child stdout/stderr to the process.
// moduleRoot/relPath should already be post-re-root (nearest go.mod + rebased path).
func runGoInModule(moduleRoot, subcmd, relPath string) error {
	return runGoInModuleTo(moduleRoot, subcmd, relPath, os.Stdout, os.Stderr)
}

func runGoInModuleTo(moduleRoot, subcmd, relPath string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	cmd := exec.Command("go", subcmd, relPath)
	cmd.Dir = moduleRoot
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	// Env is inherited (GOBIN, PATH, etc.) so installs land in the caller's bin dir.
	return cmd.Run()
}
