package unwind

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

type requireEntry struct {
	Path    string
	Version string
}

func parseRequiresTolerant(goModPath string) ([]requireEntry, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	if f, err := modfile.Parse(goModPath, data, nil); err == nil {
		var out []requireEntry
		for _, req := range f.Require {
			out = append(out, requireEntry{Path: req.Mod.Path, Version: req.Mod.Version})
		}
		return out, nil
	}
	var out []requireEntry
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
					out = append(out, requireEntry{Path: path, Version: ver})
				}
			}
			continue
		}
		if line == ")" {
			inBlock = false
			continue
		}
		if path, ver, ok := splitRequirePathVersion(line); ok {
			out = append(out, requireEntry{Path: path, Version: ver})
		}
	}
	return out, sc.Err()
}

func splitRequirePathVersion(s string) (path, version string, ok bool) {
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return "", "", false
	}
	path = fields[0]
	version = fields[1]
	if path == "" || version == "" {
		return "", "", false
	}
	if version[0] != 'v' && !strings.Contains(version, "-") {
		return "", "", false
	}
	return path, version, true
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
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("no module path in %s", filepath.Join(moduleRoot, "go.mod"))
}

func genArgsHasFlag(genArgs []string, flag string) bool {
	for _, a := range genArgs {
		name := a
		if j := strings.IndexByte(a, '='); j >= 0 {
			name = a[:j]
		}
		if name == flag {
			return true
		}
	}
	return false
}
