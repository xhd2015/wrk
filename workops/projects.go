package workops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// projectsFile is the on-disk projects.json schema (compatible with wrk storage).
type projectsFile struct {
	Version  int             `json:"version"`
	Projects []projectRecord `json:"projects"`
}

type projectRecord struct {
	Path      string `json:"path"`
	AddedAt   string `json:"added_at"`
	Source    string `json:"source"`
	OriginURL string `json:"origin_url,omitempty"`
}

// ListProjects reads registered main paths from {wrkHome}/projects.json.
// Empty wrkHome uses default wrk-home resolution (WRK_HOME or ~/.wrk).
func ListProjects(wrkHome string) ([]Project, error) {
	home := wrkHome
	if home == "" {
		var err error
		home, err = resolveDefaultWrkHome()
		if err != nil {
			return nil, err
		}
	}
	absHome, err := filepath.Abs(home)
	if err != nil {
		return nil, fmt.Errorf("resolve wrk home: %w", err)
	}

	data, err := os.ReadFile(filepath.Join(absHome, "projects.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read projects.json: %w", err)
	}

	var pf projectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parse projects.json: %w", err)
	}

	out := make([]Project, 0, len(pf.Projects))
	for _, p := range pf.Projects {
		path, nerr := normalizeAbs(p.Path)
		if nerr != nil {
			path = filepath.Clean(p.Path)
		}
		out = append(out, Project{
			Path:      path,
			OriginURL: p.OriginURL,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out, nil
}
