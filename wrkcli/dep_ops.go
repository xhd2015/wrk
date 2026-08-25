package wrkcli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/mod/modfile"
)

// runDepReplace implements wrk --dep-replace <dir>… [--dry-run].
// Git: every go.mod on CollectStackInventory, gated by existing require or
// replace (self never rewritten). Not git: nearest go.mod, D7 write even
// without require. After replaces: versioned tidy via tidyDepUpdateConsumer
// (same as --dep-update; vendor/ skips). Dry-run validates every dir before
// any banner. Apply is fail-fast: a later bad arg leaves prior writes.
func runDepReplace(workDir string, paths []string, dryRun bool, ctx *invocationContext) error {
	if len(paths) == 0 {
		return fmt.Errorf("wrk: --dep-replace requires a directory or --undo")
	}
	deps := make([]depReplaceDep, 0, len(paths))
	for _, p := range paths {
		dep, err := resolveDepReplaceArg(p)
		if err != nil {
			return err
		}
		deps = append(deps, dep)
	}
	return applyStackAbsoluteReplace(workDir, deps, stackReplaceOpts{
		DryRun: dryRun,
		Quiet:  false,
		WithGo: withGoFromCtx(ctx),
	})
}

// runDepReplaceUndo implements wrk --dep-replace --undo [<dir>…] [--dry-run].
// Requires git. For each stack consumer go.mod, drops replace OldPaths present
// in the working tree but absent from HEAD's go.mod (introduced since HEAD).
// Does not rewrite other go.mod content or put back HEAD NewPaths for existing
// OldPaths. Optional dirs filter to those module paths. Then versioned tidy
// once per affected module (vendor/ skips). Empty plan → soft success.
func runDepReplaceUndo(workDir string, paths []string, dryRun bool, ctx *invocationContext) error {
	cwd, err := absAgainstProcessCwd(workDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve cwd: %w", err)
	}
	cwd = storage.NormalizePath(cwd)
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("wrk: --dep-replace --undo requires git HEAD")
	}

	var filter map[string]struct{}
	if len(paths) > 0 {
		filter = make(map[string]struct{}, len(paths))
		for _, p := range paths {
			dep, err := resolveDepReplaceArg(p)
			if err != nil {
				return err
			}
			filter[dep.modulePath] = struct{}{}
		}
	}

	consumers, err := collectDepUpdateConsumers(cwd)
	if err != nil {
		return err
	}

	tree, err := buildDepReplaceUndoTree(cwd, consumers, filter)
	if err != nil {
		return err
	}
	return applyDepReplaceUndoTree(tree, stackReplaceOpts{
		DryRun: dryRun,
		WithGo: withGoFromCtx(ctx),
	})
}

type depReplaceUndoAction struct {
	modulePath string
	newPath    string // working-tree NewPath (display only)
}

type depReplaceUndoModule struct {
	Path   string
	ModDir string
	Drops  []depReplaceUndoAction
}

type depReplaceUndoCheckout struct {
	Path    string
	Label   string
	Modules []depReplaceUndoModule
}

func buildDepReplaceUndoTree(cwd string, consumers []depUpdateConsumer, filter map[string]struct{}) ([]depReplaceUndoCheckout, error) {
	type builder struct {
		path    string
		label   string
		modIdx  map[string]int
		modules []depReplaceUndoModule
	}
	var order []string
	byCheckout := make(map[string]*builder)

	for _, c := range consumers {
		headOld, err := headGoModReplaceOldPaths(c.Checkout, c.ModDir)
		if err != nil {
			return nil, err
		}
		var drops []depReplaceUndoAction
		for _, r := range c.Replaces {
			if r.OldPath == "" {
				continue
			}
			if filter != nil {
				if _, ok := filter[r.OldPath]; !ok {
					continue
				}
			}
			if _, ok := headOld[r.OldPath]; ok {
				continue
			}
			drops = append(drops, depReplaceUndoAction{
				modulePath: r.OldPath,
				newPath:    r.NewPath,
			})
		}
		if len(drops) == 0 {
			continue
		}
		sort.Slice(drops, func(i, j int) bool {
			return drops[i].modulePath < drops[j].modulePath
		})
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
			b.modules = append(b.modules, depReplaceUndoModule{
				Path:   c.Path,
				ModDir: c.ModDir,
			})
		}
		mi := b.modIdx[c.ModDir]
		b.modules[mi].Drops = append(b.modules[mi].Drops, drops...)
	}

	out := make([]depReplaceUndoCheckout, 0, len(order))
	for _, ck := range order {
		b := byCheckout[ck]
		out = append(out, depReplaceUndoCheckout{
			Path:    b.path,
			Label:   b.label,
			Modules: b.modules,
		})
	}
	return out, nil
}

// headGoModReplaceOldPaths returns replace OldPaths from HEAD's go.mod for the
// module at modDir under checkout. Missing HEAD blob → empty set (all WT
// replaces count as introduced).
func headGoModReplaceOldPaths(checkout, modDir string) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	goModPath := filepath.Join(modDir, "go.mod")
	rel, err := filepath.Rel(checkout, goModPath)
	if err != nil {
		return nil, fmt.Errorf("wrk: rel go.mod under checkout %s: %w", checkout, err)
	}
	if strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("wrk: go.mod %s is outside checkout %s", goModPath, checkout)
	}
	blob, err := gitOutputDir(checkout, "show", "HEAD:"+filepath.ToSlash(rel))
	if err != nil {
		// Untracked / missing at HEAD → treat as no replaces on base.
		return out, nil
	}
	f, err := modfile.Parse(goModPath, []byte(blob), nil)
	if err != nil {
		f, err = modfile.ParseLax(goModPath, []byte(blob), nil)
		if err != nil {
			return nil, fmt.Errorf("wrk: parse HEAD go.mod for %s: %w", goModPath, err)
		}
	}
	for _, r := range f.Replace {
		if r.Old.Path != "" {
			out[r.Old.Path] = struct{}{}
		}
	}
	return out, nil
}

func countDepReplaceUndoTree(tree []depReplaceUndoCheckout) (drops, modules, checkouts int) {
	for _, co := range tree {
		modCount := 0
		for _, mod := range co.Modules {
			if len(mod.Drops) == 0 {
				continue
			}
			modCount++
			modules++
			drops += len(mod.Drops)
		}
		if modCount > 0 {
			checkouts++
		}
	}
	return
}

func applyDepReplaceUndoTree(tree []depReplaceUndoCheckout, opts stackReplaceOpts) error {
	nDrops, nMods, nCheckouts := countDepReplaceUndoTree(tree)
	if nDrops == 0 {
		fmt.Println("dep-replace: nothing to undo")
		return nil
	}
	dryRun := opts.DryRun
	if dryRun {
		fmt.Println("==== dep-replace --undo (dry-run) ====")
	} else {
		fmt.Println("==== dep-replace --undo ====")
	}
	fmt.Println()
	for _, co := range tree {
		fmt.Printf("  checkout  %s\n", co.Label)
		for _, mod := range co.Modules {
			if len(mod.Drops) == 0 {
				continue
			}
			fmt.Printf("    module  %s\n", mod.Path)
			for _, act := range mod.Drops {
				if !dryRun {
					editOpts := &commands.GoModEditOptions{Dir: mod.ModDir, Stderr: false, Stdout: false}
					if err := commands.GoModDropReplace(act.modulePath, editOpts); err != nil {
						return fmt.Errorf("wrk: %w", err)
					}
				}
				if dryRun {
					fmt.Printf("      would: drop  %s => %s\n", act.modulePath, act.newPath)
					continue
				}
				fmt.Printf("      drop  %s => %s\n", act.modulePath, act.newPath)
			}
			if err := tidyDepUpdateConsumer(mod.ModDir, mod.Path, dryRun, false, opts.WithGo); err != nil {
				return err
			}
		}
	}
	fmt.Println()
	if dryRun {
		fmt.Printf("dep-replace: would undo %d replaces in %d modules in %d checkouts\n", nDrops, nMods, nCheckouts)
		return nil
	}
	fmt.Printf("dep-replace: undid %d replaces in %d modules in %d checkouts\n", nDrops, nMods, nCheckouts)
	return nil
}

// stackReplaceOpts controls applyStackAbsoluteReplace reporting and tidy.
type stackReplaceOpts struct {
	DryRun bool
	Quiet  bool // bring: no banner/tree/tidy lines
	WithGo withgo.ResolveOptions
}

// applyStackAbsoluteReplace is the shared absolute-replace + versioned-tidy
// core for --dep-replace and --bring. Consumer set = collectDepUpdateConsumers
// (unwind stack when git; nearest go.mod otherwise). Git writes are gated by
// require or existing replace; self never rewritten. When an existing local
// filesystem replace already resolves to the intended absDir (relative or
// absolute), that module is left untouched (no write, no tidy). Zero gated
// consumers on git → hard error; gated-but-all-equivalent → success no-op.
// Tidy uses tidyDepUpdateConsumer once per module that actually received a
// replace write.
func applyStackAbsoluteReplace(workDir string, deps []depReplaceDep, opts stackReplaceOpts) error {
	if len(deps) == 0 {
		return fmt.Errorf("wrk: no deps to replace")
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
	tree, err := buildDepReplaceTree(cwd, consumers, deps, git)
	if err != nil {
		return err
	}
	return applyDepReplaceTree(tree, deps, opts)
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

// consumerReplaceAlreadyEquivalent reports whether c already has a local
// filesystem replace for depPath whose New resolves to the same directory as
// absDir. Relative and absolute New forms both count; prefer leaving them alone.
func consumerReplaceAlreadyEquivalent(c depUpdateConsumer, depPath, absDir string) bool {
	want := storage.NormalizePath(absDir)
	if want == "" {
		return false
	}
	for _, r := range c.Replaces {
		if r.OldPath != depPath {
			continue
		}
		if !isLocalFilesystemReplace(r.NewPath, r.NewVersion) {
			continue
		}
		resolved, err := resolveLocalReplacePath(c.ModDir, r.NewPath)
		if err != nil {
			continue
		}
		if storage.NormalizePath(resolved) == want {
			return true
		}
	}
	return false
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
	gatedAny := false

	for _, c := range consumers {
		var acts []depReplaceAction
		for _, dep := range deps {
			if !consumerReplaceGated(c, dep.modulePath, git) {
				continue
			}
			gatedAny = true
			if consumerReplaceAlreadyEquivalent(c, dep.modulePath, dep.absDir) {
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
	if git && !gatedAny {
		path := ""
		if len(deps) > 0 {
			path = deps[0].modulePath
		}
		return nil, fmt.Errorf("wrk: no stack consumer to replace %s", path)
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

// applyDepReplaceTree writes (or plans) absolute replaces then versioned tidy
// once per affected module. Quiet suppresses all stdout (bring).
func applyDepReplaceTree(tree []depReplaceCheckout, deps []depReplaceDep, opts stackReplaceOpts) error {
	dryRun := opts.DryRun
	quiet := opts.Quiet
	if !quiet {
		if dryRun {
			fmt.Println("==== dep-replace (dry-run) ====")
		} else {
			fmt.Println("==== dep-replace ====")
		}
		for _, dep := range deps {
			fmt.Printf("dep  %s => %s\n", dep.modulePath, dep.absDir)
		}
		fmt.Println()
	}
	for _, co := range tree {
		if !quiet {
			fmt.Printf("  checkout  %s\n", co.Label)
		}
		for _, mod := range co.Modules {
			if !quiet {
				fmt.Printf("    module  %s\n", mod.Path)
			}
			for _, act := range mod.Replaces {
				if !dryRun {
					if _, _, err := replace.ReplaceIn(mod.ModDir, act.absDir); err != nil {
						return fmt.Errorf("wrk: %w", err)
					}
				}
				if quiet {
					continue
				}
				if dryRun {
					fmt.Printf("      would: replace  %s => %s\n", act.modulePath, act.absDir)
					continue
				}
				fmt.Printf("      replace  %s => %s\n", act.modulePath, act.absDir)
			}
			if len(mod.Replaces) == 0 {
				continue
			}
			if err := tidyDepUpdateConsumer(mod.ModDir, mod.Path, dryRun, quiet, opts.WithGo); err != nil {
				return err
			}
		}
	}
	if quiet {
		return nil
	}
	n, c := countDepReplaceTree(tree)
	// Blank before summary only when a checkout body was printed; empty tree
	// already has the blank after the dep headers.
	if len(tree) > 0 {
		fmt.Println()
	}
	if dryRun {
		fmt.Printf("dep-replace: would replace in %d modules in %d checkouts\n", n, c)
		return nil
	}
	fmt.Printf("dep-replace: replaced in %d modules in %d checkouts\n", n, c)
	return nil
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
	n, s, c := countDepUpdateTree(tree)
	fmt.Println()
	printDirModeSummary(n, s, c, dryRun)
	return nil
}

func printDepUpdateDepHeader(modulePath, version, tag string) {
	if tag != "" {
		fmt.Printf("dep  %s -> %s  (tag %s)\n", modulePath, version, tag)
		return
	}
	fmt.Printf("dep  %s -> %s\n", modulePath, version)
}

func printDirModeSummary(modules, skipped, checkouts int, dryRun bool) {
	if dryRun {
		if skipped > 0 {
			fmt.Printf("dep-update: would update %d modules, skipped %d in %d checkouts\n", modules, skipped, checkouts)
			return
		}
		fmt.Printf("dep-update: would update %d modules in %d checkouts\n", modules, checkouts)
		return
	}
	if skipped > 0 {
		fmt.Printf("dep-update: updated %d modules, skipped %d in %d checkouts\n", modules, skipped, checkouts)
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
		var skips []depUpdateTreeSkip
		for _, dep := range deps {
			if c.Path == dep.modulePath {
				continue
			}
			old, ok := consumerRequireVersion(c, dep.modulePath)
			if !ok {
				continue
			}
			if depReplacedIntraCheckout(c, dep.modulePath) {
				skips = append(skips, depUpdateTreeSkip{
					ModulePath: dep.modulePath,
					OldVersion: old,
					NewVersion: dep.version,
				})
				continue
			}
			pins = append(pins, depUpdateTreePin{
				ModulePath: dep.modulePath,
				DepDir:     dep.absDir,
				OldVersion: old,
				NewVersion: dep.version,
			})
		}
		if len(pins) == 0 && len(skips) == 0 {
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
		b.modules[mi].Skips = append(b.modules[mi].Skips, skips...)
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

type depUpdateTreeSkip struct {
	ModulePath string
	OldVersion string
	NewVersion string
}

type depUpdateTreeModule struct {
	Path   string
	ModDir string
	Pins   []depUpdateTreePin
	Skips  []depUpdateTreeSkip
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

// depReplacedIntraCheckout reports whether consumer c has a local filesystem
// replace for depPath whose target resolves inside c's own git checkout
// (intra-module replace). Such consumers already use their own local copy of
// the dep; dep-update skips pinning them to avoid churning the dep's own
// worktree go.mod.
func depReplacedIntraCheckout(c depUpdateConsumer, depPath string) bool {
	for _, r := range c.Replaces {
		if r.OldPath != depPath {
			continue
		}
		if !isLocalFilesystemReplace(r.NewPath, r.NewVersion) {
			continue
		}
		if replaceTargetUnderToplevel(c.ModDir, r.NewPath, c.Checkout) {
			return true
		}
	}
	return false
}

func printDepUpdateBanner(dryRun bool) {
	if dryRun {
		fmt.Println("==== dep-update (dry-run) ====")
		return
	}
	fmt.Println("==== dep-update ====")
}

func countDepUpdateTree(checkouts []depUpdateTreeCheckout) (modules, skipped, nCheckouts int) {
	for _, co := range checkouts {
		if len(co.Modules) == 0 {
			continue
		}
		coHasContent := false
		for _, mod := range co.Modules {
			if len(mod.Pins) > 0 {
				modules++
				coHasContent = true
			}
			skipped += len(mod.Skips)
			if len(mod.Skips) > 0 {
				coHasContent = true
			}
		}
		if coHasContent {
			nCheckouts++
		}
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
			for _, skip := range mod.Skips {
				if dryRun {
					fmt.Printf("      would: skip  %s  (intra-module replace)\n", skip.ModulePath)
					continue
				}
				fmt.Printf("      skip  %s  (intra-module replace)\n", skip.ModulePath)
			}
			if len(mod.Pins) == 0 {
				continue
			}
			if err := tidyDepUpdateConsumer(mod.ModDir, mod.Path, dryRun, false, withGo); err != nil {
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

// tidyDepUpdateConsumer is the shared versioned tidy helper for --dep-update,
// --dep-replace, and --bring. vendor/ directory → skip tidy (never go mod
// vendor). Else versioned withgo.ModuleGoLine + withgo.Run. Quiet suppresses
// tree lines (bring). Does not use goModTidy (pin-locals/unwind).
func tidyDepUpdateConsumer(modDir, modulePath string, dryRun, quiet bool, resolveOpts withgo.ResolveOptions) error {
	vendor := filepath.Join(modDir, "vendor")
	if fi, err := os.Stat(vendor); err == nil && fi.IsDir() {
		if !quiet {
			if dryRun {
				fmt.Println("      would: skip tidy  (vendor/)")
			} else {
				fmt.Println("      skip tidy  (vendor/)")
			}
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
		if !quiet {
			printDepUpdateTidyPlan(toolchain)
		}
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
	if !quiet {
		fmt.Println("      go mod tidy ok")
	}
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
