package wrkcli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// runDepReplace implements wrk --dep-replace <dir>… [--dry-run].
// Git: every go.mod on CollectStackInventory, gated by existing require or
// replace (self never rewritten). Not git: nearest go.mod, D7 write even
// without require. No tidy. Dry-run validates every dir before any banner.
// Apply is fail-fast: a later bad arg leaves prior writes.
func runDepReplace(workDir string, paths []string, dryRun bool) error {
	if len(paths) == 0 {
		return fmt.Errorf("wrk: --dep-replace requires a directory")
	}
	cwd, err := absAgainstProcessCwd(workDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve cwd: %w", err)
	}
	cwd = storage.NormalizePath(cwd)
	git := worktree.IsInsideWorkTree(cwd)

	consumers, err := collectDepUpdateConsumers(cwd)
	if err != nil {
		return err
	}

	if dryRun {
		deps := make([]depReplaceDep, 0, len(paths))
		for _, p := range paths {
			dep, err := resolveDepReplaceArg(p)
			if err != nil {
				return err
			}
			deps = append(deps, dep)
		}
		tree, err := buildDepReplaceTree(cwd, consumers, deps, git)
		if err != nil {
			return err
		}
		printDepReplaceReport(tree, deps, true)
		return nil
	}

	deps := make([]depReplaceDep, 0, len(paths))
	for _, p := range paths {
		dep, err := resolveDepReplaceArg(p)
		if err != nil {
			return err
		}
		targets := gatedReplaceConsumers(consumers, dep.modulePath, git)
		if git && len(targets) == 0 {
			return fmt.Errorf("wrk: no stack consumer to replace %s", dep.modulePath)
		}
		for _, c := range targets {
			if _, _, err := replace.ReplaceIn(c.ModDir, dep.absDir); err != nil {
				return fmt.Errorf("wrk: %w", err)
			}
		}
		deps = append(deps, dep)
	}
	tree, err := buildDepReplaceTree(cwd, consumers, deps, git)
	if err != nil {
		return err
	}
	printDepReplaceReport(tree, deps, false)
	return nil
}

type depReplaceDep struct {
	modulePath string
	absDir     string
}

type depReplaceAction struct {
	modulePath string
	absDir     string
}

type depReplaceModule struct {
	Path     string
	ModDir   string
	Replaces []depReplaceAction
}

type depReplaceCheckout struct {
	Path    string
	Label   string
	Modules []depReplaceModule
}

func resolveDepReplaceArg(p string) (depReplaceDep, error) {
	absDep, err := absAgainstProcessCwd(p)
	if err != nil {
		return depReplaceDep{}, fmt.Errorf("wrk: resolve %s: %w", p, err)
	}
	modPath, absDir, err := resolveDepModuleForReplace(absDep)
	if err != nil {
		return depReplaceDep{}, fmt.Errorf("wrk: %w", err)
	}
	return depReplaceDep{modulePath: modPath, absDir: absDir}, nil
}

func consumerReplaceGated(c depUpdateConsumer, depPath string, git bool) bool {
	if c.Path == depPath {
		return false
	}
	if !git {
		return true
	}
	if _, ok := consumerRequireVersion(c, depPath); ok {
		return true
	}
	for _, r := range c.Replaces {
		if r.OldPath == depPath {
			return true
		}
	}
	return false
}

func gatedReplaceConsumers(consumers []depUpdateConsumer, depPath string, git bool) []depUpdateConsumer {
	var out []depUpdateConsumer
	for _, c := range consumers {
		if consumerReplaceGated(c, depPath, git) {
			out = append(out, c)
		}
	}
	return out
}

func buildDepReplaceTree(cwd string, consumers []depUpdateConsumer, deps []depReplaceDep, git bool) ([]depReplaceCheckout, error) {
	type builder struct {
		path    string
		label   string
		modIdx  map[string]int
		modules []depReplaceModule
	}
	var order []string
	byCheckout := make(map[string]*builder)

	for _, c := range consumers {
		var acts []depReplaceAction
		for _, dep := range deps {
			if !consumerReplaceGated(c, dep.modulePath, git) {
				continue
			}
			acts = append(acts, depReplaceAction{
				modulePath: dep.modulePath,
				absDir:     dep.absDir,
			})
		}
		if len(acts) == 0 {
			continue
		}
		ck := c.Checkout
		b, ok := byCheckout[ck]
		if !ok {
			b = &builder{
				path:   ck,
				label:  statusDirLine(cwd, ck),
				modIdx: make(map[string]int),
			}
			byCheckout[ck] = b
			order = append(order, ck)
		}
		if _, exists := b.modIdx[c.ModDir]; !exists {
			b.modIdx[c.ModDir] = len(b.modules)
			b.modules = append(b.modules, depReplaceModule{
				Path:   c.Path,
				ModDir: c.ModDir,
			})
		}
		mi := b.modIdx[c.ModDir]
		b.modules[mi].Replaces = append(b.modules[mi].Replaces, acts...)
	}

	out := make([]depReplaceCheckout, 0, len(order))
	for _, ck := range order {
		b := byCheckout[ck]
		out = append(out, depReplaceCheckout{
			Path:    b.path,
			Label:   b.label,
			Modules: b.modules,
		})
	}
	if git {
		n, _ := countDepReplaceTree(out)
		if n == 0 {
			path := ""
			if len(deps) > 0 {
				path = deps[0].modulePath
			}
			return nil, fmt.Errorf("wrk: no stack consumer to replace %s", path)
		}
	}
	return out, nil
}

func countDepReplaceTree(checkouts []depReplaceCheckout) (modules, nCheckouts int) {
	for _, co := range checkouts {
		if len(co.Modules) == 0 {
			continue
		}
		nCheckouts++
		modules += len(co.Modules)
	}
	return
}

func printDepReplaceReport(tree []depReplaceCheckout, deps []depReplaceDep, dryRun bool) {
	if dryRun {
		fmt.Println("==== dep-replace (dry-run) ====")
	} else {
		fmt.Println("==== dep-replace ====")
	}
	for _, dep := range deps {
		fmt.Printf("dep  %s => %s\n", dep.modulePath, dep.absDir)
	}
	fmt.Println()
	for _, co := range tree {
		fmt.Printf("  checkout  %s\n", co.Label)
		for _, mod := range co.Modules {
			fmt.Printf("    module  %s\n", mod.Path)
			for _, act := range mod.Replaces {
				if dryRun {
					fmt.Printf("      would: replace  %s => %s\n", act.modulePath, act.absDir)
					continue
				}
				fmt.Printf("      replace  %s => %s\n", act.modulePath, act.absDir)
			}
		}
	}
	n, c := countDepReplaceTree(tree)
	fmt.Println()
	if dryRun {
		fmt.Printf("dep-replace: would replace in %d modules in %d checkouts\n", n, c)
		return
	}
	fmt.Printf("dep-replace: replaced in %d modules in %d checkouts\n", n, c)
}

// runDepUpdate implements wrk --dep-update <dir>… [--dry-run] (dir mode).
// Resolves each dep's module path + latest tag, then pins every scanned module
// on the unwind stack (CollectStackInventory) that already requires that path.
// Not git → nearest go.mod. Self (consumer.Path == dep.Path) is never pinned.
// After pins: versioned tidy via withgo unless vendor/. Multi-arg fail-fast:
// every dir arg is validated before any banner/print.
func runDepUpdate(workDir string, paths []string, dryRun bool, ctx *invocationContext) error {
	if len(paths) == 0 {
		return fmt.Errorf("wrk: --dep-update requires a directory or --all")
	}
	cwd, err := absAgainstProcessCwd(workDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve cwd: %w", err)
	}
	cwd = storage.NormalizePath(cwd)

	consumers, err := collectDepUpdateConsumers(cwd)
	if err != nil {
		return err
	}

	// Preflight every dir-mode arg before any banner/print.
	deps := make([]dirModeDep, 0, len(paths))
	for _, p := range paths {
		absDep, err := absAgainstProcessCwd(p)
		if err != nil {
			return fmt.Errorf("wrk: resolve %s: %w", p, err)
		}
		if st, err := os.Stat(absDep); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("wrk: no such dir: %s", absDep)
			}
			return fmt.Errorf("wrk: resolve %s: %w", p, err)
		} else if !st.IsDir() {
			return fmt.Errorf("wrk: no such dir: %s", absDep)
		}
		if fi, err := os.Stat(filepath.Join(absDep, "go.mod")); err != nil || fi.IsDir() {
			return fmt.Errorf("wrk: not a go module: %s", absDep)
		}
		probe, err := update.Pin(update.PinOptions{
			ConsumerDir: absDep,
			DepDir:      absDep,
			DryRun:      true,
		})
		if err != nil {
			return fmt.Errorf("wrk: %w", err)
		}
		deps = append(deps, dirModeDep{
			absDir:     absDep,
			modulePath: probe.ModulePath,
			version:    probe.Version,
			tag:        probe.Tag,
		})
	}

	// Zero requirers on the whole stack is a hard error (no banner).
	for _, dep := range deps {
		matched := 0
		for _, c := range consumers {
			if c.Path == dep.modulePath {
				continue
			}
			if _, ok := consumerRequireVersion(c, dep.modulePath); ok {
				matched++
			}
		}
		if matched == 0 {
			return fmt.Errorf("wrk: no module requires %s", dep.modulePath)
		}
	}

	tree := buildDirModeTree(cwd, consumers, deps)

	printDepUpdateBanner(dryRun)
	for _, dep := range deps {
		printDepUpdateDepHeader(dep.modulePath, dep.version, dep.tag)
	}
	fmt.Println()
	if err := applyDepUpdateTree(tree, dryRun, withGoFromCtx(ctx)); err != nil {
		return err
	}
	n, c := countDepUpdateTree(tree)
	fmt.Println()
	printDirModeSummary(n, c, dryRun)
	return nil
}

func printDepUpdateDepHeader(modulePath, version, tag string) {
	if tag != "" {
		fmt.Printf("dep  %s -> %s  (tag %s)\n", modulePath, version, tag)
		return
	}
	fmt.Printf("dep  %s -> %s\n", modulePath, version)
}

func printDirModeSummary(modules, checkouts int, dryRun bool) {
	if dryRun {
		fmt.Printf("dep-update: would update %d modules in %d checkouts\n", modules, checkouts)
		return
	}
	fmt.Printf("dep-update: updated %d modules in %d checkouts\n", modules, checkouts)
}

type dirModeDep struct {
	absDir     string
	modulePath string
	version    string
	tag        string
}

func buildDirModeTree(cwd string, consumers []depUpdateConsumer, deps []dirModeDep) []depUpdateTreeCheckout {
	type builder struct {
		path    string
		label   string
		modIdx  map[string]int
		modules []depUpdateTreeModule
	}
	var order []string
	byCheckout := make(map[string]*builder)

	for _, c := range consumers {
		var pins []depUpdateTreePin
		for _, dep := range deps {
			if c.Path == dep.modulePath {
				continue
			}
			old, ok := consumerRequireVersion(c, dep.modulePath)
			if !ok {
				continue
			}
			pins = append(pins, depUpdateTreePin{
				ModulePath: dep.modulePath,
				DepDir:     dep.absDir,
				OldVersion: old,
				NewVersion: dep.version,
			})
		}
		if len(pins) == 0 {
			continue
		}
		ck := c.Checkout
		b, ok := byCheckout[ck]
		if !ok {
			b = &builder{
				path:   ck,
				label:  statusDirLine(cwd, ck),
				modIdx: make(map[string]int),
			}
			byCheckout[ck] = b
			order = append(order, ck)
		}
		if _, exists := b.modIdx[c.ModDir]; !exists {
			b.modIdx[c.ModDir] = len(b.modules)
			b.modules = append(b.modules, depUpdateTreeModule{
				Path:   c.Path,
				ModDir: c.ModDir,
			})
		}
		mi := b.modIdx[c.ModDir]
		b.modules[mi].Pins = append(b.modules[mi].Pins, pins...)
	}

	out := make([]depUpdateTreeCheckout, 0, len(order))
	for _, ck := range order {
		b := byCheckout[ck]
		out = append(out, depUpdateTreeCheckout{
			Path:    b.path,
			Label:   b.label,
			Modules: b.modules,
		})
	}
	return out
}

func moduleDirFromScan(root string, sm scan.Module) string {
	modDir := root
	if sm.Dir != "" && sm.Dir != "." {
		modDir = filepath.Join(root, filepath.FromSlash(sm.Dir))
	}
	return storage.NormalizePath(modDir)
}

// depUpdateConsumer is one go.mod on the dir-mode / --all consumer set.
type depUpdateConsumer struct {
	Path     string
	ModDir   string
	Checkout string
	Requires []scan.ModuleRequire
	Replaces []scan.ModuleReplace
}

type depUpdateTreePin struct {
	ModulePath string
	DepDir     string
	OldVersion string
	NewVersion string
}

type depUpdateTreeModule struct {
	Path   string
	ModDir string
	Pins   []depUpdateTreePin
}

type depUpdateTreeCheckout struct {
	Path    string
	Label   string
	Modules []depUpdateTreeModule
}

// collectDepUpdateConsumers is the dir-mode / --all consumer set: every go.mod
// under every CollectStackInventory member Path. Not git → nearest go.mod.
func collectDepUpdateConsumers(cwd string) ([]depUpdateConsumer, error) {
	cwd = storage.NormalizePath(cwd)
	if worktree.IsInsideWorkTree(cwd) {
		inv, err := CollectStackInventory(cwd)
		if err != nil {
			if strings.HasPrefix(err.Error(), "wrk:") {
				return nil, err
			}
			return nil, fmt.Errorf("wrk: %w", err)
		}
		var out []depUpdateConsumer
		seen := make(map[string]struct{})
		for _, mem := range inv.Members {
			checkout := storage.NormalizePath(mem.Path)
			mods, err := scanDepUpdateCheckout(checkout)
			if err != nil {
				return nil, err
			}
			for _, c := range mods {
				if _, ok := seen[c.ModDir]; ok {
					continue
				}
				seen[c.ModDir] = struct{}{}
				c.Checkout = checkout
				out = append(out, c)
			}
		}
		return out, nil
	}
	modDir, err := findModuleRootWalking(cwd)
	if err != nil {
		return nil, fmt.Errorf("wrk: %w", err)
	}
	modDir = storage.NormalizePath(modDir)
	mods, err := scanDepUpdateCheckout(modDir)
	if err != nil {
		return nil, err
	}
	for i := range mods {
		mods[i].Checkout = modDir
	}
	return mods, nil
}

func scanDepUpdateCheckout(root string) ([]depUpdateConsumer, error) {
	scanned, err := scan.Scan(root, scan.Options{})
	if err != nil {
		return nil, fmt.Errorf("wrk: scan modules under %s: %w", root, err)
	}
	var out []depUpdateConsumer
	for _, sm := range scanned {
		if sm.Path == "" {
			continue
		}
		modDir := moduleDirFromScan(root, sm)
		reqs := sm.Requires
		if len(reqs) == 0 {
			if tol, err := parseRequiresTolerant(filepath.Join(modDir, "go.mod")); err == nil && len(tol) > 0 {
				for _, r := range tol {
					reqs = append(reqs, scan.ModuleRequire{Path: r.Path, Version: r.Version})
				}
			}
		}
		out = append(out, depUpdateConsumer{
			Path:     sm.Path,
			ModDir:   modDir,
			Requires: reqs,
			Replaces: sm.Replaces,
		})
	}
	return out, nil
}

func consumerRequireVersion(c depUpdateConsumer, depPath string) (string, bool) {
	for _, r := range c.Requires {
		if r.Path == depPath {
			return r.Version, true
		}
	}
	return "", false
}

func printDepUpdateBanner(dryRun bool) {
	if dryRun {
		fmt.Println("==== dep-update (dry-run) ====")
		return
	}
	fmt.Println("==== dep-update ====")
}

func countDepUpdateTree(checkouts []depUpdateTreeCheckout) (modules, nCheckouts int) {
	for _, co := range checkouts {
		if len(co.Modules) == 0 {
			continue
		}
		nCheckouts++
		modules += len(co.Modules)
	}
	return
}

func applyDepUpdateTree(checkouts []depUpdateTreeCheckout, dryRun bool, withGo withgo.ResolveOptions) error {
	for _, co := range checkouts {
		fmt.Printf("  checkout  %s\n", co.Label)
		for _, mod := range co.Modules {
			fmt.Printf("    module  %s\n", mod.Path)
			for _, pin := range mod.Pins {
				if dryRun {
					fmt.Printf("      would: pin  %s  %s -> %s\n", pin.ModulePath, pin.OldVersion, pin.NewVersion)
					continue
				}
				_, err := update.Pin(update.PinOptions{
					ConsumerDir: mod.ModDir,
					DepDir:      pin.DepDir,
					Version:     pin.NewVersion,
					DryRun:      false,
				})
				if err != nil {
					return fmt.Errorf("wrk: %w", err)
				}
				fmt.Printf("      pin  %s  %s -> %s\n", pin.ModulePath, pin.OldVersion, pin.NewVersion)
			}
			if err := tidyDepUpdateConsumer(mod.ModDir, mod.Path, dryRun, withGo); err != nil {
				return err
			}
		}
	}
	return nil
}

func withGoFromCtx(ctx *invocationContext) withgo.ResolveOptions {
	if ctx == nil {
		return withgo.ResolveOptions{}
	}
	return ctx.withGo
}

// resolveDepUpdateWithGo fills production defaults when InstallDir is empty:
// $HOME/installed via userHomeDir (Capture HOME) and Download:true.
func resolveDepUpdateWithGo(opts withgo.ResolveOptions) (withgo.ResolveOptions, error) {
	if opts.InstallDir != "" {
		return opts, nil
	}
	home, err := userHomeDir()
	if err != nil {
		return opts, fmt.Errorf("wrk: resolve home for go install dir: %w", err)
	}
	opts.InstallDir = filepath.Join(home, "installed")
	opts.Download = true
	return opts, nil
}

// tidyDepUpdateConsumer is the dep-update-only tidy helper (dir-mode and --all).
// vendor/ directory → skip tidy (never go mod vendor). Else versioned
// withgo.ModuleGoLine + withgo.Run. Does not use goModTidy (bring/pin-locals/unwind).
func tidyDepUpdateConsumer(modDir, modulePath string, dryRun bool, resolveOpts withgo.ResolveOptions) error {
	vendor := filepath.Join(modDir, "vendor")
	if fi, err := os.Stat(vendor); err == nil && fi.IsDir() {
		if dryRun {
			fmt.Println("      would: skip tidy  (vendor/)")
		} else {
			fmt.Println("      skip tidy  (vendor/)")
		}
		return nil
	}
	absDir, err := absAgainstProcessCwd(modDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve module dir %s: %w", modDir, err)
	}
	ver, err := withgo.ModuleGoLine(absDir)
	if err != nil {
		return fmt.Errorf("wrk: go line in %s: %w", absDir, err)
	}
	resolveOpts, err = resolveDepUpdateWithGo(resolveOpts)
	if err != nil {
		return err
	}
	toolchain := planDepUpdateGoToolchain(ver, resolveOpts)
	if dryRun {
		printDepUpdateTidyPlan(toolchain)
		return nil
	}

	var buf bytes.Buffer
	execOpts := withgo.ExecOptions{Dir: absDir}
	if invocationVerbose {
		logDepUpdateGoCommand(absDir, toolchain)
		mw := io.MultiWriter(os.Stderr, &buf)
		execOpts.Stdout = mw
		execOpts.Stderr = mw
	} else {
		execOpts.Stdout = &buf
		execOpts.Stderr = &buf
	}
	if err := withgo.Run(ver, []string{"go", "mod", "tidy"}, resolveOpts, execOpts); err != nil {
		msg := strings.TrimSpace(buf.String())
		if msg != "" {
			return fmt.Errorf("wrk: go mod tidy in %s: %w\n%s", absDir, err, msg)
		}
		return fmt.Errorf("wrk: go mod tidy in %s: %w", absDir, err)
	}
	fmt.Println("      go mod tidy ok")
	return nil
}

type depUpdateGoToolchain struct {
	goroot   string
	override bool
}

// planDepUpdateGoToolchain predicts the same pinned GOROOT that withgo.Run
// will use, without checking for or downloading the SDK. This lets dry-run
// remain mutation-free while accurately describing a future tidy command.
func planDepUpdateGoToolchain(version string, opts withgo.ResolveOptions) depUpdateGoToolchain {
	goroot := withgo.TargetGoroot(version, opts)
	defaultGoroot, err := defaultGoGOROOT()
	return depUpdateGoToolchain{
		goroot:   goroot,
		override: err == nil && !sameCleanPath(goroot, defaultGoroot),
	}
}

func defaultGoGOROOT() (string, error) {
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func sameCleanPath(a, b string) bool {
	if a == "" || b == "" {
		return a == b
	}
	if resolved, err := filepath.EvalSymlinks(a); err == nil {
		a = resolved
	}
	if resolved, err := filepath.EvalSymlinks(b); err == nil {
		b = resolved
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func printDepUpdateTidyPlan(toolchain depUpdateGoToolchain) {
	if !toolchain.override {
		fmt.Println("      would: go mod tidy")
		return
	}
	fmt.Printf("      would: go mod tidy  (go=%s; GOROOT=%s)\n", filepath.Base(toolchain.goroot), toolchain.goroot)
}

func logDepUpdateGoCommand(absDir string, toolchain depUpdateGoToolchain) {
	if !toolchain.override {
		logGoCommand([]string{"-C", absDir, "mod", "tidy"})
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	goBin := filepath.Join(toolchain.goroot, "bin", "go")
	fmt.Fprintf(os.Stderr, "[%s] $ GOROOT=%s %s -C %s mod tidy\n", ts, toolchain.goroot, goBin, absDir)
}

// resolveDepModuleForReplace resolves module path + absolute dep dir without writing.
// Mirrors replace.ReplaceIn validation (exists + go module).
func resolveDepModuleForReplace(dir string) (modulePath, absDir string, err error) {
	if dir == "" {
		return "", "", fmt.Errorf("requires dir")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("no such dir: %s", dir)
	}
	absDir, err = absAgainstProcessCwd(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %v", err)
	}
	modInfo, err := resolve.GetModuleInfo(absDir)
	if err != nil {
		// Opaque library errors (e.g. "resolve go mod: … exit status 1") still mean
		// the path is not a usable Go module for --dep-replace.
		return "", "", fmt.Errorf("not a go module: %s", absDir)
	}
	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", absDir)
	}
	return modInfo.Module.Path, absDir, nil
}
