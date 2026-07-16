package wrkcli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/mod/modfile"
)

// Inventory is the in-process model of registered projects, their modules, and
// soft-skipped registry paths.
type Inventory struct {
	Projects     []ProjectEntry
	SkippedPaths []string
}

// ProjectEntry is one existing registered main-repo path and its Go modules.
type ProjectEntry struct {
	Path    string
	Modules []ModuleEntry
}

// ModuleEntry describes one go.mod under a project.
type ModuleEntry struct {
	Dir      string // relative to project Path; "." for root
	Path     string // module path from go.mod
	Requires []RequireEntry
	Replaces []ReplaceEntry
}

// RequireEntry is a single require directive from go.mod.
type RequireEntry struct {
	Path    string
	Version string
}

// ReplaceEntry is a single replace directive from go.mod.
type ReplaceEntry struct {
	OldPath    string
	NewPath    string
	NewVersion string
}

// Edge is a require from a consumer module to a dependency module path whose
// owner project is known in the inventory.
type Edge struct {
	ConsumerProject string
	ConsumerModule  string
	DepPath         string
	DepVersion      string
	OwnerProject    string
}

// SourceReleasesResult maps modules under a source main repo to their latest
// numeric release tags (or lists modules with no such tag).
type SourceReleasesResult struct {
	Releases []SourceRelease
	Missing  []string
}

// SourceRelease is one module's resolved git tag and go require version.
type SourceRelease struct {
	ModulePath string
	Tag        string // full git tag, e.g. "v1.2.3" or "sub/v0.1.0"
	Version    string // go require version, e.g. "v1.2.3" or "v0.1.0"
}

// BuildInventory loads WRK_HOME projects, soft-skips missing paths, scans
// modules, and builds ownership.
func BuildInventory(wrkHome string) (Inventory, error) {
	var inv Inventory
	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return inv, err
	}

	for _, projectPath := range paths {
		if _, err := os.Stat(projectPath); err != nil {
			// Soft-skip registry paths that are missing on disk.
			if os.IsNotExist(err) {
				inv.SkippedPaths = append(inv.SkippedPaths, projectPath)
				continue
			}
			return inv, err
		}

		modules, err := scan.Scan(projectPath, scan.Options{})
		if err != nil {
			return inv, err
		}

		entry := ProjectEntry{Path: projectPath}
		for _, m := range modules {
			mod := ModuleEntry{
				Dir:  m.Dir,
				Path: m.Path,
			}
			for _, req := range m.Requires {
				mod.Requires = append(mod.Requires, RequireEntry{
					Path:    req.Path,
					Version: req.Version,
				})
			}
			for _, repl := range m.Replaces {
				mod.Replaces = append(mod.Replaces, ReplaceEntry{
					OldPath:    repl.OldPath,
					NewPath:    repl.NewPath,
					NewVersion: repl.NewVersion,
				})
			}
			// scan.Scan uses modfile.Parse, which rejects the entire require
			// block when any require has an invalid major/path version pair
			// (e.g. example.com/external v9.9.9). Fall back to a tolerant
			// require parser so valid sibling requires still surface as edges.
			if len(mod.Requires) == 0 {
				modDir := projectPath
				if m.Dir != "" && m.Dir != "." {
					modDir = filepath.Join(projectPath, filepath.FromSlash(m.Dir))
				}
				if reqs, err := parseRequiresTolerant(filepath.Join(modDir, "go.mod")); err == nil && len(reqs) > 0 {
					mod.Requires = reqs
				}
			}
			entry.Modules = append(entry.Modules, mod)
		}
		inv.Projects = append(inv.Projects, entry)
	}
	return inv, nil
}

// parseRequiresTolerant extracts require path@version pairs from go.mod even
// when strict modfile validation would reject some versions.
func parseRequiresTolerant(goModPath string) ([]RequireEntry, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	if f, err := modfile.Parse(goModPath, data, nil); err == nil {
		var out []RequireEntry
		for _, req := range f.Require {
			out = append(out, RequireEntry{Path: req.Mod.Path, Version: req.Mod.Version})
		}
		return out, nil
	}
	// Line-oriented fallback: handle `require path ver` and parenthesized blocks.
	var out []RequireEntry
	inBlock := false
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		if !inBlock {
			if line == "require (" {
				inBlock = true
				continue
			}
			if strings.HasPrefix(line, "require ") && !strings.HasPrefix(line, "require (") {
				rest := strings.TrimSpace(strings.TrimPrefix(line, "require "))
				if path, ver, ok := splitRequirePathVersion(rest); ok {
					out = append(out, RequireEntry{Path: path, Version: ver})
				}
			}
			continue
		}
		if line == ")" {
			inBlock = false
			continue
		}
		if path, ver, ok := splitRequirePathVersion(line); ok {
			out = append(out, RequireEntry{Path: path, Version: ver})
		}
	}
	return out, sc.Err()
}

func splitRequirePathVersion(s string) (path, version string, ok bool) {
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return "", "", false
	}
	// Drop trailing comments already stripped; ignore //indirect markers as fields.
	path = fields[0]
	version = fields[1]
	if path == "" || version == "" {
		return "", "", false
	}
	// Version tokens start with v or are pseudo-versions / revisions.
	if version[0] != 'v' && !strings.Contains(version, "-") {
		return "", "", false
	}
	return path, version, true
}

// FindOwner returns the registered project path that owns modulePath.
func (inv Inventory) FindOwner(modulePath string) (projectPath string, ok bool) {
	if modulePath == "" {
		return "", false
	}
	for _, p := range inv.Projects {
		for _, m := range p.Modules {
			if m.Path == modulePath {
				return p.Path, true
			}
		}
	}
	return "", false
}

// CrossEdges returns require edges where consumer and owner projects both
// known and differ.
func (inv Inventory) CrossEdges() []Edge {
	return inv.collectEdges(true)
}

// IntraEdges returns require edges where consumer and owner projects both
// known and equal.
func (inv Inventory) IntraEdges() []Edge {
	return inv.collectEdges(false)
}

func (inv Inventory) collectEdges(cross bool) []Edge {
	var edges []Edge
	for _, p := range inv.Projects {
		for _, m := range p.Modules {
			for _, req := range m.Requires {
				if req.Path == "" {
					continue
				}
				owner, ok := inv.FindOwner(req.Path)
				if !ok {
					// Unknown owners (external modules): neither list.
					continue
				}
				sameProject := owner == p.Path
				if cross == sameProject {
					// cross=true wants differ; cross=false wants equal.
					continue
				}
				edges = append(edges, Edge{
					ConsumerProject: p.Path,
					ConsumerModule:  m.Path,
					DepPath:         req.Path,
					DepVersion:      req.Version,
					OwnerProject:    owner,
				})
			}
		}
	}
	return edges
}

// ResolveSourceReleases scans sourceMain for modules and maps numeric release
// tags to go require versions. Modules without a numeric tag are listed in
// Missing; overall success is still returned when any/all modules are missing.
func ResolveSourceReleases(sourceMain string) (SourceReleasesResult, error) {
	var result SourceReleasesResult
	sourceMain = storage.NormalizePath(sourceMain)

	modules, err := scan.Scan(sourceMain, scan.Options{})
	if err != nil {
		return result, err
	}

	for _, m := range modules {
		if m.Path == "" {
			continue
		}
		modDir := sourceMain
		if m.Dir != "" && m.Dir != "." {
			modDir = filepath.Join(sourceMain, filepath.FromSlash(m.Dir))
		}

		versionPrefix, err := update.CalculateVersionPrefix(modDir, m.Path)
		if err != nil {
			result.Missing = append(result.Missing, m.Path)
			continue
		}
		tag, err := update.GetLatestVersionTag(modDir, versionPrefix)
		if err != nil {
			result.Missing = append(result.Missing, m.Path)
			continue
		}
		version := update.StripVersionPrefix(versionPrefix, tag)
		result.Releases = append(result.Releases, SourceRelease{
			ModulePath: m.Path,
			Tag:        tag,
			Version:    version,
		})
	}
	return result, nil
}
