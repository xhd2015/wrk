package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// DepUpdateAllAction is one planned pin under --dep-update --all.
type DepUpdateAllAction struct {
	ConsumerModDir string
	ConsumerPath   string
	CheckoutPath   string // stack member Path containing the consumer
	DepModule      string
	DepDir         string // owner module dir on registered main path
	OldVersion     string
	NewVersion     string
	Tag            string
}

// DepUpdateAllPlan is the pure plan for wrk --dep-update --all.
type DepUpdateAllPlan struct {
	Actions   []DepUpdateAllAction
	Already   int
	Skipped   int
	Warnings  []string // may include warning: prefix
	Checkouts int      // stack member count (used when there are no pin actions)
}

// inventoryModule locates an inventory-owned module on a registered main path.
type inventoryModule struct {
	ProjectPath string
	ModDir      string
	Path        string
}

// PlanDepUpdateAll builds pin actions for inventory-owned requires on the
// unwind stack (CollectStackInventory). External (non-inventory) requires are
// silent. Same-checkout local filesystem replaces and no-tag owners count as
// skipped. --all still requires a git repo.
func PlanDepUpdateAll(workDir, wrkHome string) (*DepUpdateAllPlan, error) {
	cwd, err := absAgainstProcessCwd(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	cwd = storage.NormalizePath(cwd)
	if !worktree.IsInsideWorkTree(cwd) {
		return nil, fmt.Errorf("wrk: %s is not a git repository", cwd)
	}

	stack, err := CollectStackInventory(cwd)
	if err != nil {
		if strings.HasPrefix(err.Error(), "wrk:") {
			return nil, err
		}
		return nil, fmt.Errorf("wrk: %w", err)
	}

	inv, err := BuildInventory(wrkHome)
	if err != nil {
		return nil, err
	}

	plan := &DepUpdateAllPlan{}
	for _, p := range inv.SkippedPaths {
		plan.Warnings = append(plan.Warnings,
			fmt.Sprintf("warning: project path does not exist: %s", p))
	}

	// modulePath → owner module on registered main path (first match).
	owners := make(map[string]inventoryModule)
	for _, proj := range inv.Projects {
		projPath := storage.NormalizePath(proj.Path)
		for _, m := range proj.Modules {
			if m.Path == "" {
				continue
			}
			if _, exists := owners[m.Path]; exists {
				continue
			}
			modDir := projPath
			if m.Dir != "" && m.Dir != "." {
				modDir = filepath.Join(projPath, filepath.FromSlash(m.Dir))
			}
			owners[m.Path] = inventoryModule{
				ProjectPath: projPath,
				ModDir:      storage.NormalizePath(modDir),
				Path:        m.Path,
			}
		}
	}

	// Cache latest tag resolution per owner module dir.
	type releaseInfo struct {
		Tag     string
		Version string
		OK      bool
		ErrMsg  string
	}
	releaseCache := make(map[string]releaseInfo) // modDir → release
	resolveRelease := func(modDir, modPath string) releaseInfo {
		if cached, ok := releaseCache[modDir]; ok {
			return cached
		}
		info := releaseInfo{}
		versionPrefix, err := update.CalculateVersionPrefix(modDir, modPath)
		if err != nil {
			info.ErrMsg = err.Error()
			releaseCache[modDir] = info
			return info
		}
		tag, err := update.GetLatestVersionTag(modDir, versionPrefix)
		if err != nil {
			info.ErrMsg = err.Error()
			releaseCache[modDir] = info
			return info
		}
		version := update.StripVersionPrefix(versionPrefix, tag)
		info.Tag = tag
		info.Version = version
		info.OK = version != ""
		releaseCache[modDir] = info
		return info
	}

	consumers, err := collectDepUpdateConsumers(cwd)
	if err != nil {
		return nil, err
	}

	plan.Checkouts = len(stack.Members)

	for _, c := range consumers {
		// Stable require order.
		type reqItem struct {
			Path    string
			Version string
		}
		var reqs []reqItem
		seenReq := make(map[string]struct{})
		for _, r := range c.Requires {
			if r.Path == "" || r.Path == c.Path {
				continue
			}
			if _, ok := seenReq[r.Path]; ok {
				continue
			}
			seenReq[r.Path] = struct{}{}
			reqs = append(reqs, reqItem{Path: r.Path, Version: r.Version})
		}
		sort.Slice(reqs, func(i, j int) bool { return reqs[i].Path < reqs[j].Path })

		// Map replace OldPath → NewPath/NewVersion for this consumer.
		type replInfo struct {
			NewPath    string
			NewVersion string
		}
		replByOld := make(map[string]replInfo)
		for _, repl := range c.Replaces {
			if repl.OldPath == "" {
				continue
			}
			replByOld[repl.OldPath] = replInfo{NewPath: repl.NewPath, NewVersion: repl.NewVersion}
		}

		for _, req := range reqs {
			// Same-toplevel + local filesystem replace → skip (count); leave as-is.
			if repl, ok := replByOld[req.Path]; ok && isLocalFilesystemReplace(repl.NewPath, repl.NewVersion) {
				if replaceTargetUnderToplevel(c.ModDir, repl.NewPath, c.Checkout) {
					plan.Skipped++
					continue
				}
			}

			owner, owned := owners[req.Path]
			if !owned {
				// External / non-inventory: silent skip (do not count).
				continue
			}

			rel := resolveRelease(owner.ModDir, owner.Path)
			if !rel.OK {
				plan.Skipped++
				msg := fmt.Sprintf("warning: no version tag for %s", req.Path)
				if rel.ErrMsg != "" {
					msg = fmt.Sprintf("warning: no version tag for %s: %s", req.Path, rel.ErrMsg)
				}
				plan.Warnings = append(plan.Warnings, msg)
				continue
			}

			if req.Version == rel.Version {
				plan.Already++
				continue
			}

			plan.Actions = append(plan.Actions, DepUpdateAllAction{
				ConsumerModDir: c.ModDir,
				ConsumerPath:   c.Path,
				CheckoutPath:   c.Checkout,
				DepModule:      req.Path,
				DepDir:         owner.ModDir,
				OldVersion:     req.Version,
				NewVersion:     rel.Version,
				Tag:            rel.Tag,
			})
		}
	}

	return plan, nil
}

// replaceTargetUnderToplevel reports whether a local replace NewPath resolves
// to the same git checkout as consumerToplevel. Nested independent repos
// (e.g. ./external/kool) are other stack members, not intra-checkout skips.
func replaceTargetUnderToplevel(consumerModDir, newPath, consumerToplevel string) bool {
	if newPath == "" {
		return false
	}
	target := newPath
	if !filepath.IsAbs(target) {
		target = filepath.Join(consumerModDir, filepath.FromSlash(newPath))
	}
	target = storage.NormalizePath(target)
	top := storage.NormalizePath(consumerToplevel)
	if worktree.IsInsideWorkTree(target) {
		targetTop, err := worktree.ShowToplevel(target)
		if err == nil {
			return storage.NormalizePath(targetTop) == top
		}
	}
	rel, err := filepath.Rel(top, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != "" && !strings.HasPrefix(rel, ".."))
}

// runDepUpdateAll implements wrk --dep-update --all [--dry-run].
// Consumer set = CollectStackInventory(cwd). Pins inventory-owned outdated
// requires, then tidyDepUpdateConsumer once per affected consumer. No commit/build.
func runDepUpdateAll(workDir, wrkHome string, dryRun bool, ctx *invocationContext) error {
	plan, err := PlanDepUpdateAll(workDir, wrkHome)
	if err != nil {
		return err
	}

	for _, w := range plan.Warnings {
		line := w
		if !strings.HasPrefix(line, "warning:") {
			line = "warning: " + line
		}
		fmt.Fprintln(os.Stderr, FormatStderrWarning(line))
	}

	// No pin actions: already-up-to-date form (apply wording even with --dry-run).
	if len(plan.Actions) == 0 {
		printDepUpdateBanner(false)
		fmt.Println("dep-update: already up to date")
		c := plan.Checkouts
		if c <= 0 {
			c = 1
		}
		fmt.Printf("dep-update: updated 0, already %d, skipped %d in %d checkouts\n", plan.Already, plan.Skipped, c)
		return nil
	}

	cwd, err := absAgainstProcessCwd(workDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve cwd: %w", err)
	}
	tree := treeFromAllPlan(plan, storage.NormalizePath(cwd))
	printDepUpdateBanner(dryRun)
	fmt.Println()
	if err := applyDepUpdateTree(tree, dryRun, withGoFromCtx(ctx)); err != nil {
		return err
	}
	_, c := countDepUpdateTree(tree)
	fmt.Println()
	n := len(plan.Actions)
	if dryRun {
		fmt.Printf("dep-update: would update %d, already %d, skipped %d in %d checkouts\n", n, plan.Already, plan.Skipped, c)
	} else {
		fmt.Printf("dep-update: updated %d, already %d, skipped %d in %d checkouts\n", n, plan.Already, plan.Skipped, c)
	}
	return nil
}

func treeFromAllPlan(plan *DepUpdateAllPlan, cwd string) []depUpdateTreeCheckout {
	type builder struct {
		path    string
		label   string
		modIdx  map[string]int
		modules []depUpdateTreeModule
	}
	var order []string
	byCheckout := make(map[string]*builder)
	for _, a := range plan.Actions {
		ck := a.CheckoutPath
		if ck == "" {
			ck = a.ConsumerModDir
		}
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
		mi, exists := b.modIdx[a.ConsumerModDir]
		if !exists {
			mi = len(b.modules)
			b.modIdx[a.ConsumerModDir] = mi
			b.modules = append(b.modules, depUpdateTreeModule{
				Path:   a.ConsumerPath,
				ModDir: a.ConsumerModDir,
			})
		}
		b.modules[mi].Pins = append(b.modules[mi].Pins, depUpdateTreePin{
			ModulePath: a.DepModule,
			DepDir:     a.DepDir,
			OldVersion: a.OldVersion,
			NewVersion: a.NewVersion,
		})
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
