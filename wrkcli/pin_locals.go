package wrkcli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// Pin action kinds for PlanPinLocals.
const (
	pinKindAdd     = "add"
	pinKindRewrite = "rewrite"
	pinKindAlready = "already"
)

// PinLocalsAction is one planned relative replace for a consumer→dep edge.
type PinLocalsAction struct {
	ConsumerModDir string
	ConsumerPath   string
	DepPath        string
	DepModDir      string
	RelPath        string // ./ or ../ slash form
	Kind           string // add | rewrite | already
}

// PinLocalsPlan is the pure plan for wrk --pin-locals.
type PinLocalsPlan struct {
	Actions  []PinLocalsAction
	Warnings []string // soft multi-owner etc. (may include warning: prefix)
}

type pinModuleOwner struct {
	ModDir   string // absolute module directory
	Checkout string // stack member checkout path containing the module
}

// PlanPinLocals builds relative-replace actions from the unwind stack inventory.
// Inventory = CollectStackInventory only (no WRK_HOME project universe).
// Wanted set per consumer = require paths ∪ replace OldPaths that have a stack owner.
func PlanPinLocals(workDir string) (*PinLocalsPlan, error) {
	inv, err := CollectStackInventory(workDir)
	if err != nil {
		return nil, err
	}

	// modulePath → owners (may be multi when same path appears in multiple places)
	owners := make(map[string][]pinModuleOwner)
	// consumer modules found under any stack checkout
	type consumerMod struct {
		Path     string
		ModDir   string
		Checkout string
		Requires []scan.ModuleRequire
		Replaces []scan.ModuleReplace
	}
	var consumers []consumerMod
	seenConsumer := make(map[string]struct{})

	for _, mem := range inv.Members {
		checkout := storage.NormalizePath(mem.Path)
		scanned, scanErr := scan.Scan(checkout, scan.Options{})
		if scanErr != nil {
			return nil, scanErr
		}
		for _, sm := range scanned {
			if sm.Path == "" {
				continue
			}
			modDir := checkout
			if sm.Dir != "" && sm.Dir != "." {
				modDir = filepath.Join(checkout, filepath.FromSlash(sm.Dir))
			}
			modDir = storage.NormalizePath(modDir)
			owners[sm.Path] = append(owners[sm.Path], pinModuleOwner{
				ModDir:   modDir,
				Checkout: checkout,
			})
			key := modDir
			if _, ok := seenConsumer[key]; ok {
				continue
			}
			seenConsumer[key] = struct{}{}
			consumers = append(consumers, consumerMod{
				Path:     sm.Path,
				ModDir:   modDir,
				Checkout: checkout,
				Requires: sm.Requires,
				Replaces: sm.Replaces,
			})
		}
	}

	// Stable owner order per path for multi-owner selection.
	for path, list := range owners {
		sort.Slice(list, func(i, j int) bool {
			if list[i].ModDir != list[j].ModDir {
				return list[i].ModDir < list[j].ModDir
			}
			return list[i].Checkout < list[j].Checkout
		})
		owners[path] = list
	}
	sort.Slice(consumers, func(i, j int) bool {
		if consumers[i].Path != consumers[j].Path {
			return consumers[i].Path < consumers[j].Path
		}
		return consumers[i].ModDir < consumers[j].ModDir
	})

	plan := &PinLocalsPlan{}
	// Copy inventory soft warnings (optional; not required by asserts).
	plan.Warnings = append(plan.Warnings, inv.Warnings...)

	for _, c := range consumers {
		wanted := make(map[string]struct{})
		for _, req := range c.Requires {
			if req.Path != "" {
				wanted[req.Path] = struct{}{}
			}
		}
		existingNew := make(map[string]string) // oldPath → NewPath
		for _, repl := range c.Replaces {
			if repl.OldPath == "" {
				continue
			}
			wanted[repl.OldPath] = struct{}{}
			existingNew[repl.OldPath] = repl.NewPath
		}

		// Stable dep order.
		depPaths := make([]string, 0, len(wanted))
		for p := range wanted {
			depPaths = append(depPaths, p)
		}
		sort.Strings(depPaths)

		for _, depPath := range depPaths {
			// Never pin self.
			if depPath == c.Path {
				continue
			}
			ownList, ok := owners[depPath]
			if !ok || len(ownList) == 0 {
				continue // no stack owner
			}
			owner, warn := pickPinOwner(ownList, c.ModDir, c.Checkout)
			if warn != "" {
				plan.Warnings = append(plan.Warnings, warn)
			}
			rel, relErr := relativeReplacePath(c.ModDir, owner.ModDir)
			if relErr != nil {
				return nil, fmt.Errorf("relative path %s -> %s: %w", c.ModDir, owner.ModDir, relErr)
			}
			kind := pinKindAdd
			if cur, has := existingNew[depPath]; has {
				if replaceAlreadyRelative(c.ModDir, cur, rel, owner.ModDir) {
					kind = pinKindAlready
				} else {
					kind = pinKindRewrite
				}
			}
			plan.Actions = append(plan.Actions, PinLocalsAction{
				ConsumerModDir: c.ModDir,
				ConsumerPath:   c.Path,
				DepPath:        depPath,
				DepModDir:      owner.ModDir,
				RelPath:        rel,
				Kind:           kind,
			})
		}
	}

	return plan, nil
}

// pickPinOwner implements D4: same checkout → under consumer checkout → stable first + warning.
func pickPinOwner(owners []pinModuleOwner, consumerModDir, consumerCheckout string) (pinModuleOwner, string) {
	if len(owners) == 1 {
		return owners[0], ""
	}
	var same []pinModuleOwner
	for _, o := range owners {
		if o.Checkout == consumerCheckout {
			same = append(same, o)
		}
	}
	if len(same) == 1 {
		return same[0], ""
	}
	if len(same) > 1 {
		// Prefer owner under consumer module dir if nested, else first stable.
		var underMod []pinModuleOwner
		for _, o := range same {
			if pathIsUnderOrEqual(o.ModDir, consumerModDir) {
				underMod = append(underMod, o)
			}
		}
		if len(underMod) >= 1 {
			return underMod[0], ""
		}
		return same[0], ""
	}
	var underCheckout []pinModuleOwner
	for _, o := range owners {
		if pathIsUnderOrEqual(o.ModDir, consumerCheckout) || pathIsUnderOrEqual(o.Checkout, consumerCheckout) {
			underCheckout = append(underCheckout, o)
		}
	}
	if len(underCheckout) == 1 {
		return underCheckout[0], ""
	}
	if len(underCheckout) > 1 {
		return underCheckout[0], ""
	}
	// Stable first + warning.
	first := owners[0]
	return first, fmt.Sprintf(
		"warning: multiple stack owners for module path; using %s",
		first.ModDir,
	)
}

// relativeReplacePath returns a slash-form relative path suitable for go.mod replace NewPath.
func relativeReplacePath(fromModDir, toModDir string) (string, error) {
	from, err := filepath.Abs(fromModDir)
	if err != nil {
		return "", err
	}
	to, err := filepath.Abs(toModDir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(from, to)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return "./", nil
	}
	if strings.HasPrefix(rel, "../") || strings.HasPrefix(rel, "./") {
		return rel, nil
	}
	// filepath.Rel yields "tools" for a child; go.mod wants "./tools".
	return "./" + rel, nil
}

// replaceAlreadyRelative reports whether existing NewPath is already the desired relative pin.
func replaceAlreadyRelative(consumerModDir, existingNew, wantedRel, depModDir string) bool {
	if existingNew == "" {
		return false
	}
	// Exact relative match (normalize slash).
	if filepath.ToSlash(existingNew) == wantedRel {
		return true
	}
	// Absolute same target is NOT already — must rewrite to relative (D2).
	if filepath.IsAbs(existingNew) {
		return false
	}
	// Relative form that resolves to the same dep dir.
	resolved, err := resolveLocalReplacePath(consumerModDir, existingNew)
	if err != nil {
		return false
	}
	resolved = storage.NormalizePath(resolved)
	want := storage.NormalizePath(depModDir)
	if resolved != want {
		return false
	}
	// Only treat as already if it's relative (./ or ../).
	p := filepath.ToSlash(existingNew)
	return strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../")
}

// runPinLocals implements wrk --pin-locals [--dry-run] [--color|--no-color].
func runPinLocals(workDir string, dryRun bool, colorFlag bool) error {
	_ = colorFlag // stderr color uses SetForceStderrColor from run(); stdout color optional
	plan, err := PlanPinLocals(workDir)
	if err != nil {
		return err
	}

	// Emit soft plan warnings (multi-owner).
	for _, w := range plan.Warnings {
		line := w
		if !strings.HasPrefix(line, "warning:") {
			line = "warning: " + line
		}
		fmt.Fprintln(os.Stderr, FormatStderrWarning(line))
	}

	work := make([]PinLocalsAction, 0, len(plan.Actions))
	for _, a := range plan.Actions {
		if a.Kind == pinKindAdd || a.Kind == pinKindRewrite {
			work = append(work, a)
		}
	}

	if len(work) == 0 {
		// Keep "already" / "up to date" wording; also emit zero summary so leaves
		// that check pin-locals: applied 0 … pass when stdout contains pin-locals:.
		fmt.Println("pin-locals: already up to date")
		fmt.Println("pin-locals: applied 0, tidy ok 0, tidy failed 0")
		return nil
	}

	if dryRun {
		for _, a := range work {
			fmt.Printf("would: pin-local %s <- %s => %s\n", a.ConsumerPath, a.DepPath, a.RelPath)
		}
		return nil
	}

	// Group work by consumer module for one tidy per modified consumer.
	type consumerWork struct {
		ModDir string
		Path   string
		Acts   []PinLocalsAction
	}
	order := make([]string, 0)
	byMod := make(map[string]*consumerWork)
	for _, a := range work {
		cw, ok := byMod[a.ConsumerModDir]
		if !ok {
			cw = &consumerWork{ModDir: a.ConsumerModDir, Path: a.ConsumerPath}
			byMod[a.ConsumerModDir] = cw
			order = append(order, a.ConsumerModDir)
		}
		cw.Acts = append(cw.Acts, a)
	}

	applied := 0
	tidyOK := 0
	tidyFailed := 0

	for _, modDir := range order {
		cw := byMod[modDir]
		for _, a := range cw.Acts {
			opts := &commands.GoModEditOptions{Dir: a.ConsumerModDir, Stderr: false, Stdout: false}
			if err := commands.GoModEditReplace(a.DepPath, a.RelPath, opts); err != nil {
				return fmt.Errorf("pin-local replace %s => %s in %s: %w", a.DepPath, a.RelPath, a.ConsumerModDir, err)
			}
			fmt.Printf("pin-local %s <- %s => %s\n", a.ConsumerPath, a.DepPath, a.RelPath)
			applied++
		}
		if err := goModTidyForPinLocals(cw.ModDir); err != nil {
			tidyFailed++
			msg := fmt.Sprintf("warning: go mod tidy in %s: %v", cw.ModDir, err)
			fmt.Fprintln(os.Stderr, FormatStderrWarning(msg))
			continue
		}
		tidyOK++
	}

	fmt.Printf("pin-locals: applied %d, tidy ok %d, tidy failed %d\n", applied, tidyOK, tidyFailed)
	return nil
}

// goModTidyForPinLocals runs go mod tidy, then soft-fails if any pre-tidy require
// without a local filesystem replace cannot be resolved (e.g. GOPROXY=off and no
// stack owner). Modern go mod tidy drops unused requires without network, so a
// plain tidy success alone would hide unresolvable declared deps; pin-locals
// treats that as a soft tidy failure (warning + continue).
func goModTidyForPinLocals(dir string) error {
	before, err := listGoModRequires(dir)
	if err != nil {
		return goModTidy(dir)
	}
	// Snapshot which paths already have local replaces before tidy mutates go.mod.
	needsResolve := make([]goModRequire, 0)
	for _, r := range before {
		if r.path == "" {
			continue
		}
		if goModHasLocalReplace(dir, r.path) {
			continue
		}
		needsResolve = append(needsResolve, r)
	}
	if err := goModTidy(dir); err != nil {
		return err
	}
	for _, r := range needsResolve {
		if err := probeModuleRequire(dir, r.path, r.version); err != nil {
			return fmt.Errorf("failed to execute go mod tidy: %w", err)
		}
	}
	return nil
}

// probeModuleRequire checks that path[@version] is resolvable from dir (honors GOPROXY).
func probeModuleRequire(dir, path, version string) error {
	arg := path
	if version != "" {
		arg = path + "@" + version
	}
	cmd := exec.Command("go", "list", "-m", arg)
	cmd.Dir = dir
	var buf strings.Builder
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(buf.String())
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}
