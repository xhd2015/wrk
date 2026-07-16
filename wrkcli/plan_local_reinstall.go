package wrkcli

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

// Method and Action are string type aliases so harness code can use string(it.Method).
type Method string
type Action string

const (
	MethodGoInstall    Method = "go-install"
	MethodGoRunInstall Method = "go-run-install"
	ActionInstall      Action = "install"
	ActionSkip         Action = "skip"
)

// LocalReinstallPlan is the pure discovery/filter result for local binary reinstalls.
type LocalReinstallPlan struct {
	ModuleRoot string
	ModulePath string // full module path from go.mod
	ModuleName string // basename of module path from go.mod
	BinDir     string
	Items      []PlanItem // sorted lexicographically by BinName
}

// MultiLocalReinstallPlan is the multi-module discovery/filter result for a
// shared binDir. Modules are ordered lexicographically by absolute ModuleRoot.
type MultiLocalReinstallPlan struct {
	BinDir  string
	Modules []ModuleReinstallPlan
}

// ModuleReinstallPlan is one module's contribution to a multi-module plan.
type ModuleReinstallPlan struct {
	ModuleRoot string
	ModulePath string // full module path from go.mod
	ModuleName string // basename of module path from go.mod
	RelDir     string // module root relative to scan root ("." or slash path); set by FromWorkDir
	Items      []PlanItem // sorted lexicographically by BinName
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
func PlanLocalReinstalls(moduleRoot, binDir string) (*LocalReinstallPlan, error) {
	modulePath, err := readModulePath(moduleRoot)
	if err != nil {
		return nil, err
	}
	moduleName := filepath.Base(modulePath)
	if moduleName == "" || moduleName == "." || moduleName == string(filepath.Separator) {
		return nil, fmt.Errorf("invalid module path %q", modulePath)
	}

	// Keyed by BinName; script discovery overwrites cmd (script wins).
	byName := make(map[string]PlanItem)

	if err := walkCmdMains(moduleRoot, byName); err != nil {
		return nil, err
	}
	if err := walkScriptInstalls(moduleRoot, moduleName, byName); err != nil {
		return nil, err
	}

	items := make([]PlanItem, 0, len(byName))
	for _, it := range byName {
		it.Action = binAction(binDir, it.BinName)
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].BinName < items[j].BinName
	})

	return &LocalReinstallPlan{
		ModuleRoot: moduleRoot,
		ModulePath: modulePath,
		ModuleName: moduleName,
		BinDir:     binDir,
		Items:      items,
	}, nil
}

// PlanLocalReinstallsMulti runs PlanLocalReinstalls for each module root and
// returns a multi-module plan sorted by absolute ModuleRoot path.
//
// Empty moduleRoots yields an empty Modules list and nil error.
// After per-module planning, if the same BinName has Action=install from two
// or more modules, returns a hard error naming the bin and both modules.
// Skip-only (or install×skip) duplicates across modules are allowed.
func PlanLocalReinstallsMulti(moduleRoots []string, binDir string) (*MultiLocalReinstallPlan, error) {
	if len(moduleRoots) == 0 {
		return &MultiLocalReinstallPlan{
			BinDir:  binDir,
			Modules: []ModuleReinstallPlan{},
		}, nil
	}

	modules := make([]ModuleReinstallPlan, 0, len(moduleRoots))
	for _, root := range moduleRoots {
		plan, err := PlanLocalReinstalls(root, binDir)
		if err != nil {
			return nil, err
		}
		modules = append(modules, ModuleReinstallPlan{
			ModuleRoot: plan.ModuleRoot,
			ModulePath: plan.ModulePath,
			ModuleName: plan.ModuleName,
			Items:      plan.Items,
		})
	}

	sort.Slice(modules, func(i, j int) bool {
		ai := absModuleRoot(modules[i].ModuleRoot)
		aj := absModuleRoot(modules[j].ModuleRoot)
		return ai < aj
	})

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
//
// Rules (priority):
//  1. useMain == true: workDir must be inside a git checkout; scan root is the
//     main repository path (ResolveMainRepo of ShowToplevel).
//  2. useMain == false and inside git: scan root is ShowToplevel(workDir).
//  3. Not in git: walk up from workDir looking for a go.mod; first hit is root.
//  4. No scan root: not in git and no go.mod on the walk-up → error.
func ResolveReinstallScanRoot(workDir string, useMain bool) (string, error) {
	top, err := worktree.ShowToplevel(workDir)
	if err == nil {
		if useMain {
			mainRepo, err := worktree.ResolveMainRepo(top)
			if err != nil {
				return "", fmt.Errorf("resolve main repo for reinstall scan: %w", err)
			}
			return mainRepo, nil
		}
		return top, nil
	}
	if useMain {
		return "", fmt.Errorf("resolve reinstall scan root with main: not inside a git work tree: %w", err)
	}
	// Not in git: walk up looking for go.mod (same class as single-module path).
	return findModuleRootWalking(workDir)
}

// PlanLocalReinstallsFromWorkDir resolves the scan root from workDir, discovers
// every Go module under that root via mod/scan, and builds a multi-module
// reinstall plan against binDir.
//
// Zero modules under the scan root is a hard error (message mentions go.mod).
// Each ModuleReinstallPlan.RelDir is the scan-relative module dir (scan.Module.Dir).
func PlanLocalReinstallsFromWorkDir(workDir, binDir string, useMain bool) (*MultiLocalReinstallPlan, error) {
	scanRoot, err := ResolveReinstallScanRoot(workDir, useMain)
	if err != nil {
		return nil, err
	}
	modules, err := scan.Scan(scanRoot, scan.Options{})
	if err != nil {
		return nil, fmt.Errorf("scan modules under %s: %w", scanRoot, err)
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("no go.mod modules found under %s", scanRoot)
	}
	moduleRoots := make([]string, 0, len(modules))
	// abs module root -> RelDir (scan.Module.Dir, already "." or slash path)
	relDirByAbsRoot := make(map[string]string, len(modules))
	for _, m := range modules {
		modDir := scanRoot
		if m.Dir != "." {
			modDir = filepath.Join(scanRoot, filepath.FromSlash(m.Dir))
		}
		moduleRoots = append(moduleRoots, modDir)
		relDirByAbsRoot[absModuleRoot(modDir)] = m.Dir
	}
	plan, err := PlanLocalReinstallsMulti(moduleRoots, binDir)
	if err != nil {
		return nil, err
	}
	for i := range plan.Modules {
		if rel, ok := relDirByAbsRoot[absModuleRoot(plan.Modules[i].ModuleRoot)]; ok {
			plan.Modules[i].RelDir = rel
		} else {
			// Fallback: compute RelDir from scan root.
			plan.Modules[i].RelDir = relDirFromScanRoot(scanRoot, plan.Modules[i].ModuleRoot)
		}
	}
	return plan, nil
}

// relDirFromScanRoot returns moduleRoot relative to scanRoot in slash form, or ".".
func relDirFromScanRoot(scanRoot, moduleRoot string) string {
	rel, err := filepath.Rel(absModuleRoot(scanRoot), absModuleRoot(moduleRoot))
	if err != nil {
		return moduleRoot
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return "."
	}
	return rel
}

// absModuleRoot returns an absolute path for sort keys; falls back to root on error.
func absModuleRoot(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}

// installClaim tracks one module that wants to install a given bin.
type installClaim struct {
	ModuleRoot string
	ModuleName string
}

// detectCrossModuleInstallCollisions returns an error when the same BinName has
// Action=install in two or more modules.
func detectCrossModuleInstallCollisions(modules []ModuleReinstallPlan) error {
	// binName -> first install claim(s)
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
	// Deterministic: report the lexicographically smallest colliding bin name.
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
	// Identify both modules by path and name so error substrings match either form.
	a, b := list[0], list[1]
	return fmt.Errorf(
		"bin %q claimed for install by multiple modules: %s (%s) and %s (%s)",
		bin, a.ModuleRoot, a.ModuleName, b.ModuleRoot, b.ModuleName,
	)
}

func readModulePath(moduleRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(moduleRoot, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		// module <path>  (optional trailing comment)
		if !strings.HasPrefix(line, "module ") && line != "module" {
			// First non-empty non-comment line that is not a module directive
			// means unparseable for our purposes if we never saw module.
			// Keep scanning — go.mod can theoretically have leading comments only.
			if strings.HasPrefix(line, "module\t") {
				// handled below via Fields
			} else if !strings.HasPrefix(line, "module") {
				continue
			}
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
		if len(fields) == 1 && fields[0] == "module" {
			return "", fmt.Errorf("go.mod: module directive missing path")
		}
	}
	return "", fmt.Errorf("go.mod: no module path found")
}

func walkCmdMains(moduleRoot string, byName map[string]PlanItem) error {
	cmdRoot := filepath.Join(moduleRoot, "cmd")
	info, err := os.Stat(cmdRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(cmdRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		// Skip testdata, vendor, and hidden dirs (but not the cmd root itself).
		if path != cmdRoot {
			if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}
		// Do not treat cmd root itself as a package candidate for bin naming
		// in the usual sense — still check isPackageMain for completeness;
		// BinName would be "cmd" which is unusual but allowed if package main.
		ok, err := isPackageMainDir(path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rel, err := filepath.Rel(moduleRoot, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		binName := filepath.Base(path)
		byName[binName] = PlanItem{
			BinName: binName,
			RelPath: "./" + relSlash,
			Method:  MethodGoInstall,
		}
		return nil
	})
}

func walkScriptInstalls(moduleRoot, moduleName string, byName map[string]PlanItem) error {
	scriptRoot := filepath.Join(moduleRoot, "script")
	info, err := os.Stat(scriptRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(scriptRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != scriptRoot {
			if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}
		// Only directories named "install" are script install candidates.
		if name != "install" {
			return nil
		}
		ok, err := isPackageMainDir(path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		parent := filepath.Base(filepath.Dir(path))
		binName := parent
		// bare ./script/install → parent base is "script" → use module basename
		if parent == "script" {
			// Ensure parent dir is actually the script root (not .../something/script/install
			// where something/script is not moduleRoot/script — still: parent base "script"
			// is the rule from the spec: if parent is `script` (path is `script/install`)).
			// Spec: "if parent is script (path is script/install) → BinName = ModuleName"
			// So only when the install dir's parent is exactly moduleRoot/script.
			if filepath.Clean(filepath.Dir(path)) == filepath.Clean(scriptRoot) {
				binName = moduleName
			}
		}
		rel, err := filepath.Rel(moduleRoot, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		// Script wins over cmd for same BinName.
		byName[binName] = PlanItem{
			BinName: binName,
			RelPath: "./" + relSlash,
			Method:  MethodGoRunInstall,
		}
		return nil
	})
}

// isPackageMainDir reports whether dir contains any non-test *.go file declaring package main.
func isPackageMainDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		// Parse only the package clause for speed.
		f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			// Unparseable file: skip it; directory may still have other mains.
			continue
		}
		if f.Name != nil && f.Name.Name == "main" {
			return true, nil
		}
	}
	return false, nil
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

// runReinstallLocal implements wrk --reinstall-local [--dry-run] [--main].
// dry-run prints the plan and does not run go install/run.
// Without --dry-run, installs run sequentially (continue on failure).
//
// Planning uses PlanLocalReinstallsFromWorkDir(workDir, binDir, useMain).
// useMain=false scans the worktree toplevel (or walk-up); useMain=true
// (from --main) scans the main repository of this checkout.
func runReinstallLocal(workDir string, dryRun bool, useMain bool) error {
	binDir, err := resolveLocalReinstallBinDir()
	if err != nil {
		return err
	}
	plan, err := PlanLocalReinstallsFromWorkDir(workDir, binDir, useMain)
	if err != nil {
		return err
	}
	if dryRun {
		return printMultiLocalReinstallDryRun(plan)
	}
	return executeMultiLocalReinstalls(plan)
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
func printLocalReinstallDryRun(plan *LocalReinstallPlan) error {
	nInstall, nSkip := 0, 0
	if err := printPlanItemsDryRun(plan.Items, plan.BinDir, &nInstall, &nSkip); err != nil {
		return err
	}
	fmt.Printf("would: reinstall %d binaries (%d skipped)\n", nInstall, nSkip)
	return nil
}

// printMultiLocalReinstallDryRun prints a multi-module dry-run plan.
//
// K==1: same format as single-module dry-run (no # module headers; summary
// without "across").
// K>1: for each module in plan order, "# module <ModulePath> (<RelDir>)" then
// that module's would:/skip: lines; summary ends with " across K modules".
func printMultiLocalReinstallDryRun(plan *MultiLocalReinstallPlan) error {
	k := len(plan.Modules)
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
		if err := printPlanItemsDryRun(mod.Items, plan.BinDir, &nInstall, &nSkip); err != nil {
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

// printPlanItemsDryRun writes would:/skip: lines for items; accumulates counters.
func printPlanItemsDryRun(items []PlanItem, binDir string, nInstall, nSkip *int) error {
	for _, it := range items {
		switch it.Action {
		case ActionInstall:
			*nInstall++
			switch it.Method {
			case MethodGoInstall:
				fmt.Printf("would: go install %s\n", it.RelPath)
			case MethodGoRunInstall:
				fmt.Printf("would: go run %s\n", it.RelPath)
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
// Summary: reinstalled N, skipped M, failed F. Exit 1 iff failed > 0.
func executeLocalReinstalls(plan *LocalReinstallPlan) error {
	nReinstalled, nSkip, nFailed := 0, 0, 0
	if err := executePlanItems(plan.ModuleRoot, plan.BinDir, plan.Items, &nReinstalled, &nSkip, &nFailed); err != nil {
		return err
	}
	fmt.Printf("reinstalled %d, skipped %d, failed %d\n", nReinstalled, nSkip, nFailed)
	if nFailed > 0 {
		return ExitCodeError{Code: 1}
	}
	return nil
}

// executeMultiLocalReinstalls runs installs for every module in the multi plan
// (module order, then BinName order within each module). Same progress/skip
// lines and summary as single-module execute. Continues after failures.
func executeMultiLocalReinstalls(plan *MultiLocalReinstallPlan) error {
	nReinstalled, nSkip, nFailed := 0, 0, 0
	for _, mod := range plan.Modules {
		if err := executePlanItems(mod.ModuleRoot, plan.BinDir, mod.Items, &nReinstalled, &nSkip, &nFailed); err != nil {
			return err
		}
	}
	fmt.Printf("reinstalled %d, skipped %d, failed %d\n", nReinstalled, nSkip, nFailed)
	if nFailed > 0 {
		return ExitCodeError{Code: 1}
	}
	return nil
}

// executePlanItems runs install/skip actions for one module's items.
// Unknown method/action returns a hard error (stops the plan).
func executePlanItems(moduleRoot, binDir string, items []PlanItem, nReinstalled, nSkip, nFailed *int) error {
	for _, it := range items {
		switch it.Action {
		case ActionInstall:
			var err error
			switch it.Method {
			case MethodGoInstall:
				fmt.Printf("go install %s\n", it.RelPath)
				err = runGoInModule(moduleRoot, "install", it.RelPath)
			case MethodGoRunInstall:
				fmt.Printf("go run %s\n", it.RelPath)
				err = runGoInModule(moduleRoot, "run", it.RelPath)
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
			fmt.Printf("skip: %s (not in %s)\n", it.BinName, binDir)
		default:
			return fmt.Errorf("unknown reinstall action %q for %s", it.Action, it.BinName)
		}
	}
	return nil
}

// runGoInModule runs `go <subcmd> <relPath>` with Dir=moduleRoot and inherited env
// (including GOBIN). Streams child stdout/stderr to the process.
func runGoInModule(moduleRoot, subcmd, relPath string) error {
	cmd := exec.Command("go", subcmd, relPath)
	cmd.Dir = moduleRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Env is inherited (GOBIN, PATH, etc.) so installs land in the caller's bin dir.
	return cmd.Run()
}
