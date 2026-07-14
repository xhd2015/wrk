package storage

import (
	"path/filepath"
	"sort"
)

// FindProjectsByBasename returns saved project paths whose final path component
// equals basename, sorted lexicographically.
func FindProjectsByBasename(wrkHome, basename string) ([]string, error) {
	pf, err := LoadProjects(wrkHome)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, p := range pf.Projects {
		path := NormalizePath(p.Path)
		if filepath.Base(path) == basename {
			matches = append(matches, path)
		}
	}
	sort.Strings(matches)
	return matches, nil
}