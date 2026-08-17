package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// DepUpdateAllAction is one planned pin under --dep-update --all.
type DepUpdateAllAction struct {
	ConsumerModDir string
	ConsumerPath   string
	DepModule      string
	DepDir         string // owner module dir on registered main path
	OldVersion     string
	NewVersion     string
	Tag            string
}

// DepUpdateAllPlan is the pure plan for wrk --dep-update --all.
type DepUpdateAllPlan struct {
	Actions  []DepUpdateAllAction
	Already  int
	Skipped  int
	Warnings []string // may include warning: prefix
}

// inventoryModule locates an inventory-owned module on a registered main path.
type inventoryModule struct {
	ProjectPath string
	ModDir      string
	Path        string
}

// PlanDepUpdateAll builds pin actions for inventory-owned requires under the
// git toplevel of workDir. External (non-inventory) requires are silent.
// Same-toplevel local filesystem replaces and no-tag owners count as skipped.
func PlanDepUpdateAll(workDir, wrkHome string) (*DepUpdateAllPlan, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return nil, fmt.Errorf("wrk: %s is not a git repository", cwd)
	}
	toplevel, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return nil, fmt.Errorf("wrk: resolve git toplevel: %w", err)
	}
	toplevel = storage.NormalizePath(toplevel)

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

	// Scan consumer modules under current git toplevel only.
	scanned, err := scan.Scan(toplevel, scan.Options{})
	if err != nil {
		return nil, fmt.Errorf("wrk: scan modules under %s: %w", toplevel, err)
	}

	type consumerMod struct {
		Path     string
		ModDir   string
		Requires []scan.ModuleRequire
		Replaces []scan.ModuleReplace
	}
	var consumers []consumerMod
	for _, sm := range scanned {
		if sm.Path == "" {
			continue
		}
		modDir := toplevel
		if sm.Dir != "" && sm.Dir != "." {
			modDir = filepath.Join(toplevel, filepath.FromSlash(sm.Dir))
		}
		modDir = storage.NormalizePath(modDir)
		reqs := sm.Requires
		// Tolerant parse when scan dropped requires (invalid major/path pairs).
		if len(reqs) == 0 {
			if tol, err := parseRequiresTolerant(filepath.Join(modDir, "go.mod")); err == nil && len(tol) > 0 {
				for _, r := range tol {
					reqs = append(reqs, scan.ModuleRequire{Path: r.Path, Version: r.Version})
				}
			}
		}
		consumers = append(consumers, consumerMod{
			Path:     sm.Path,
			ModDir:   modDir,
			Requires: reqs,
			Replaces: sm.Replaces,
		})
	}
	sort.Slice(consumers, func(i, j int) bool {
		if consumers[i].Path != consumers[j].Path {
			return consumers[i].Path < consumers[j].Path
		}
		return consumers[i].ModDir < consumers[j].ModDir
	})

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
				if replaceTargetUnderToplevel(c.ModDir, repl.NewPath, toplevel) {
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
// to a path under consumerToplevel.
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
	rel, err := filepath.Rel(top, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != "" && !strings.HasPrefix(rel, ".."))
}

// runDepUpdateAll implements wrk --dep-update --all [--dry-run].
// Consumer root = git toplevel of cwd. Pins inventory-owned outdated requires,
// then tidyDepUpdateConsumer once per affected consumer module. No commit/build.
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
		fmt.Println("dep-update: already up to date")
		fmt.Printf("dep-update: updated 0, already %d, skipped %d\n", plan.Already, plan.Skipped)
		return nil
	}

	// Group actions by consumer for one tidy per affected module; stream pins.
	type consumerGroup struct {
		ModDir string
		Path   string
		Acts   []DepUpdateAllAction
	}
	order := make([]string, 0)
	byMod := make(map[string]*consumerGroup)
	for _, a := range plan.Actions {
		cg, ok := byMod[a.ConsumerModDir]
		if !ok {
			cg = &consumerGroup{ModDir: a.ConsumerModDir, Path: a.ConsumerPath}
			byMod[a.ConsumerModDir] = cg
			order = append(order, a.ConsumerModDir)
		}
		cg.Acts = append(cg.Acts, a)
	}

	updated := 0
	for _, modDir := range order {
		cg := byMod[modDir]
		for _, a := range cg.Acts {
			if dryRun {
				fmt.Printf("would: dep-update %s -> %s\n", a.DepModule, a.NewVersion)
				updated++
				continue
			}
			result, err := update.Pin(update.PinOptions{
				ConsumerDir: a.ConsumerModDir,
				DepDir:      a.DepDir,
				// Force resolved inventory version so owner tag set is authoritative.
				Version: a.NewVersion,
				DryRun:  false,
			})
			if err != nil {
				return fmt.Errorf("wrk: %w", err)
			}
			tag := a.Tag
			if result.Tag != "" {
				tag = result.Tag
			}
			if tag != "" {
				fmt.Printf("dep-update %s -> %s  (tag %s)\n", result.ModulePath, result.Version, tag)
			} else {
				fmt.Printf("dep-update %s -> %s\n", result.ModulePath, result.Version)
			}
			updated++
		}
		if err := tidyDepUpdateConsumer(cg.ModDir, cg.Path, dryRun, withGoFromCtx(ctx)); err != nil {
			return err
		}
	}

	if dryRun {
		fmt.Printf("dep-update: would update %d, already %d, skipped %d\n", updated, plan.Already, plan.Skipped)
	} else {
		fmt.Printf("dep-update: updated %d, already %d, skipped %d\n", updated, plan.Already, plan.Skipped)
	}
	return nil
}
