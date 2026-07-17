package wrkcli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// plannedBump is one require version change for a consumer module.
type plannedBump struct {
	DepModule string
	OldVer    string
	NewVer    string
}

// plannedConsumerUpdate groups bumps (and optional drop-replaces) for one
// consumer module under a registered project.
type plannedConsumerUpdate struct {
	ConsumerModule string
	ModuleDir      string // absolute dir holding the consumer go.mod
	ProjectPath    string
	ProjectBase    string
	Bumps          []plannedBump
	DropReplaces   []string // dep module paths

	// Apply results (non-dry-run). BuildOK is true only when every updated
	// module in ProjectPath built successfully. CommitShort is non-empty when
	// a deps commit was created for the project.
	BuildOK       bool
	CommitShort   string
	CommitSubject string
}

// runPropagateTags plans and optionally applies consumer go.mod tag updates from
// the source project's latest numeric release tags.
//
// dry-run prints the human plan and performs no writes. Apply drops local
// replaces, edits requires to release versions, runs go mod tidy, then gates
// on go build ./... per updated module and commits go.mod/go.sum when all
// modules in a consumer project build cleanly.
func runPropagateTags(workDir, wrkHome string, dryRun bool) error {
	return runPropagateTagsWithReleases(workDir, wrkHome, dryRun, nil)
}

// runPropagateTagsWithReleases is like runPropagateTags but, when releaseOverride
// is non-nil, uses that release set instead of ResolveSourceReleases (compose
// dry-run threads planned next tags that do not exist yet).
func runPropagateTagsWithReleases(workDir, wrkHome string, dryRun bool, releaseOverride []SourceRelease) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	mainRepo, err := resolveMainRepoForWorkDir(cwd)
	if err != nil {
		return err
	}
	sourceMain := storage.NormalizePath(mainRepo)

	var sourceReleases []SourceRelease
	if releaseOverride != nil {
		sourceReleases = releaseOverride
	} else {
		releases, err := ResolveSourceReleases(sourceMain)
		if err != nil {
			return err
		}
		sourceReleases = releases.Releases
	}
	if len(sourceReleases) == 0 {
		return fmt.Errorf("wrk: no usable release tags for source modules")
	}

	inv, err := BuildInventory(wrkHome)
	if err != nil {
		return err
	}
	for _, p := range inv.SkippedPaths {
		fmt.Fprintf(os.Stderr, "warning: project path does not exist: %s\n", p)
	}

	var updates []plannedConsumerUpdate
	for _, proj := range inv.Projects {
		projPath := storage.NormalizePath(proj.Path)
		if projPath == sourceMain {
			// Intra-project / self: never a consumer of its own source releases.
			continue
		}
		for _, mod := range proj.Modules {
			var bumps []plannedBump
			// Preserve source-release order for stable dep lines.
			for _, rel := range sourceReleases {
				for _, req := range mod.Requires {
					if req.Path != rel.ModulePath {
						continue
					}
					if req.Version == rel.Version {
						continue
					}
					bumps = append(bumps, plannedBump{
						DepModule: rel.ModulePath,
						OldVer:    req.Version,
						NewVer:    rel.Version,
					})
				}
			}
			if len(bumps) == 0 {
				continue
			}
			// Local filesystem replaces for deps that would be updated.
			bumpSet := make(map[string]struct{}, len(bumps))
			for _, b := range bumps {
				bumpSet[b.DepModule] = struct{}{}
			}
			var drops []string
			seenDrop := make(map[string]struct{})
			for _, repl := range mod.Replaces {
				if _, ok := bumpSet[repl.OldPath]; !ok {
					continue
				}
				if !isLocalFilesystemReplace(repl.NewPath, repl.NewVersion) {
					continue
				}
				if _, seen := seenDrop[repl.OldPath]; seen {
					continue
				}
				seenDrop[repl.OldPath] = struct{}{}
				drops = append(drops, repl.OldPath)
			}
			// Prefer source-release order for drop lines too.
			if len(drops) > 1 {
				ordered := make([]string, 0, len(drops))
				for _, rel := range sourceReleases {
					if _, ok := seenDrop[rel.ModulePath]; ok {
						ordered = append(ordered, rel.ModulePath)
					}
				}
				drops = ordered
			}
			modDir := projPath
			if mod.Dir != "" && mod.Dir != "." {
				modDir = filepath.Join(projPath, filepath.FromSlash(mod.Dir))
			}
			updates = append(updates, plannedConsumerUpdate{
				ConsumerModule: mod.Path,
				ModuleDir:      modDir,
				ProjectPath:    projPath,
				ProjectBase:    filepath.Base(projPath),
				Bumps:          bumps,
				DropReplaces:   drops,
			})
		}
	}

	if !dryRun {
		if err := applyPropagateTags(updates); err != nil {
			return err
		}
	}
	fmt.Fprint(os.Stdout, formatPropagateTagsPlan(sourceMain, sourceReleases, updates, dryRun))
	return nil
}

// applyPropagateTags mutates consumer go.mod files for each planned update:
// drop local replaces, set require versions, go mod tidy, then go build ./...
// per module. On full project build success, stages only go.mod/go.sum under
// edited module dirs and creates one chore(deps) commit per project. Build
// failures are soft: warning on stderr, no commit, continue (overall exit 0).
func applyPropagateTags(updates []plannedConsumerUpdate) error {
	for i := range updates {
		u := &updates[i]
		opts := &commands.GoModEditOptions{Dir: u.ModuleDir, Stderr: false, Stdout: false}
		for _, dep := range u.DropReplaces {
			if err := commands.GoModDropReplace(dep, opts); err != nil {
				return fmt.Errorf("drop replace %s in %s: %w", dep, u.ModuleDir, err)
			}
		}
		for _, bump := range u.Bumps {
			if err := commands.GoModEditRequire(bump.DepModule, bump.NewVer, opts); err != nil {
				return fmt.Errorf("edit require %s@%s in %s: %w", bump.DepModule, bump.NewVer, u.ModuleDir, err)
			}
		}
		if err := commands.GoModTidy(opts); err != nil {
			return fmt.Errorf("go mod tidy in %s: %w", u.ModuleDir, err)
		}
	}

	// Group updates by project while preserving first-seen project order.
	type projGroup struct {
		path string
		idxs []int
	}
	var groups []projGroup
	groupIndex := make(map[string]int)
	for i, u := range updates {
		if gi, ok := groupIndex[u.ProjectPath]; ok {
			groups[gi].idxs = append(groups[gi].idxs, i)
			continue
		}
		groupIndex[u.ProjectPath] = len(groups)
		groups = append(groups, projGroup{path: u.ProjectPath, idxs: []int{i}})
	}

	for _, g := range groups {
		allOK := true
		for _, i := range g.idxs {
			if err := goBuildAll(updates[i].ModuleDir); err != nil {
				allOK = false
				fmt.Fprintf(os.Stderr, "warning: go build ./... failed for project %s (%s): %v\n",
					updates[i].ProjectBase, updates[i].ModuleDir, err)
				break
			}
		}
		if !allOK {
			// Leave working tree dirty; do not commit this project.
			continue
		}
		for _, i := range g.idxs {
			updates[i].BuildOK = true
		}

		subject := projectDepsCommitSubject(updates, g.idxs)
		if err := commitProjectDeps(g.path, updates, g.idxs, subject); err != nil {
			return fmt.Errorf("commit deps in %s: %w", g.path, err)
		}
		short, err := gitOutputDir(g.path, "rev-parse", "--short=7", "HEAD")
		if err != nil {
			return fmt.Errorf("rev-parse short HEAD in %s: %w", g.path, err)
		}
		short = strings.TrimSpace(short)
		for _, i := range g.idxs {
			updates[i].CommitShort = short
			updates[i].CommitSubject = subject
		}
	}
	return nil
}

// goBuildAll runs `go build ./...` with Cmd.Dir set to moduleDir.
func goBuildAll(moduleDir string) error {
	cmd := exec.Command("go", "build", "-buildvcs=false", "./...")
	cmd.Dir = moduleDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%w\n%s", err, msg)
		}
		return err
	}
	return nil
}

// projectDepsCommitSubject builds the chore(deps) commit subject for a project.
// Single-dep form: "chore(deps): bump <module> to <version>".
func projectDepsCommitSubject(updates []plannedConsumerUpdate, idxs []int) string {
	var bumps []plannedBump
	for _, i := range idxs {
		bumps = append(bumps, updates[i].Bumps...)
	}
	if len(bumps) == 1 {
		return fmt.Sprintf("chore(deps): bump %s to %s", bumps[0].DepModule, bumps[0].NewVer)
	}
	// Multi-dep: prefer first bump as representative when only one unique dep.
	seen := make(map[string]string)
	var order []string
	for _, b := range bumps {
		if _, ok := seen[b.DepModule]; !ok {
			seen[b.DepModule] = b.NewVer
			order = append(order, b.DepModule)
		}
	}
	if len(order) == 1 {
		return fmt.Sprintf("chore(deps): bump %s to %s", order[0], seen[order[0]])
	}
	return fmt.Sprintf("chore(deps): bump %d modules", len(order))
}

// commitProjectDeps stages only go.mod/go.sum under edited module dirs and commits.
func commitProjectDeps(projectPath string, updates []plannedConsumerUpdate, idxs []int, subject string) error {
	var toAdd []string
	for _, i := range idxs {
		u := updates[i]
		relDir := ""
		if u.ModuleDir != projectPath {
			rel, err := filepath.Rel(projectPath, u.ModuleDir)
			if err != nil {
				return fmt.Errorf("rel module dir: %w", err)
			}
			if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return fmt.Errorf("module dir %s outside project %s", u.ModuleDir, projectPath)
			}
			relDir = rel
		}
		modPath := "go.mod"
		sumPath := "go.sum"
		if relDir != "" {
			modPath = filepath.Join(relDir, "go.mod")
			sumPath = filepath.Join(relDir, "go.sum")
		}
		toAdd = append(toAdd, modPath)
		if _, err := os.Stat(filepath.Join(projectPath, sumPath)); err == nil {
			toAdd = append(toAdd, sumPath)
		}
	}
	if len(toAdd) == 0 {
		return fmt.Errorf("no go.mod/go.sum to stage in %s", projectPath)
	}
	args := append([]string{"add", "--"}, toAdd...)
	if err := gitRunDir(projectPath, args...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	// -q: suppress "[branch sha] subject" on stdout (product stdout is the plan report).
	// --no-verify: skip interactive/global hooks for tool-driven deps commits.
	if err := gitRunDir(projectPath, "commit", "-q", "--no-verify", "-m", subject); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

func isLocalFilesystemReplace(newPath, newVersion string) bool {
	if newPath == "" || newVersion != "" {
		return false
	}
	return strings.HasPrefix(newPath, "./") ||
		strings.HasPrefix(newPath, "../") ||
		filepath.IsAbs(newPath)
}

// formatPropagateTagsPlan renders the human plan or apply report (always ends with \n).
// When dryRun is true, lines use "would: …"; otherwise past-tense apply wording.
// On apply success, indented build/commit lines follow version arrows (before drop-replace).
func formatPropagateTagsPlan(sourceMain string, releases []SourceRelease, updates []plannedConsumerUpdate, dryRun bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "source: %s\n", sourceMain)
	for _, r := range releases {
		fmt.Fprintf(&b, "  %s  @ %s  (tag %s)\n", r.ModulePath, r.Version, r.Tag)
	}
	b.WriteByte('\n')

	bumpCount := 0
	projectSet := make(map[string]struct{})
	for _, u := range updates {
		bumpCount += len(u.Bumps)
		projectSet[u.ProjectPath] = struct{}{}

		if dryRun {
			fmt.Fprintf(&b, "would: update %s  (project %s)\n", u.ConsumerModule, u.ProjectBase)
		} else {
			fmt.Fprintf(&b, "updated %s  (project %s)\n", u.ConsumerModule, u.ProjectBase)
		}
		for _, bump := range u.Bumps {
			fmt.Fprintf(&b, "  %s  %s -> %s\n", bump.DepModule, bump.OldVer, bump.NewVer)
		}
		if !dryRun && u.BuildOK {
			fmt.Fprintf(&b, "  go build ./... ok\n")
			if u.CommitShort != "" {
				fmt.Fprintf(&b, "  committed %s  %s\n", u.CommitShort, u.CommitSubject)
			}
		}
		b.WriteByte('\n')
		for _, dep := range u.DropReplaces {
			if dryRun {
				fmt.Fprintf(&b, "would: drop replace %s  (project %s)\n", dep, u.ProjectBase)
			} else {
				fmt.Fprintf(&b, "dropped replace %s  (project %s)\n", dep, u.ProjectBase)
			}
			b.WriteByte('\n')
		}
	}

	if dryRun {
		fmt.Fprintf(&b, "would: update %d %s across %d %s\n",
			bumpCount, pluralWord(bumpCount, "module", "modules"),
			len(projectSet), pluralWord(len(projectSet), "project", "projects"),
		)
	} else {
		fmt.Fprintf(&b, "updated %d %s across %d %s\n",
			bumpCount, pluralWord(bumpCount, "module", "modules"),
			len(projectSet), pluralWord(len(projectSet), "project", "projects"),
		)
	}
	return b.String()
}
