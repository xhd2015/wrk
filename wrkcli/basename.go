package wrkcli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/term"
)

func isBasename(dir string) bool {
	if dir == "" || filepath.IsAbs(dir) {
		return false
	}
	if strings.ContainsRune(dir, filepath.Separator) {
		return false
	}
	// Also reject forward slashes on all platforms (wrk accepts Unix-style args).
	if strings.Contains(dir, "/") {
		return false
	}
	return true
}

func isCreateMode(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, repos, status bool, bring bool, reinstallLocal, tagNext, propagateTags, syncFlag, pushFlag, list, done, mergeBack, cd, mainFlag, unwind bool) bool {
	if projects || projectsDepGraph || addFlagSet || removeFlagSet || setTaskFlagSet || whereFlagSet || repos || status || cd || mainFlag || unwind {
		return false
	}
	if bring || reinstallLocal || tagNext || propagateTags || syncFlag || pushFlag || list || done || mergeBack {
		return false
	}
	return true
}

// DirHintOptions carries CLI context for reconstructing guided file-collision hints.
type DirHintOptions struct {
	RawArgs     []string
	Positionals []string
	DepMode     bool
}

// resolveDirArg resolves dir to an absolute path: Abs → stat → optional basename
// fallback via resolveBasenameFromProjects when allowBasenameFallback is true.
// Relative dirs resolve against processCwd() (Capture virtual Dir when set).
func resolveDirArg(dir string, allowBasenameFallback bool, wrkHome string, hint *DirHintOptions) (string, error) {
	absCandidate, err := absAgainstProcessCwd(dir)
	if err != nil {
		return "", fmt.Errorf("resolve dir: %w", err)
	}
	info, err := os.Stat(absCandidate)
	if err == nil {
		if info.IsDir() {
			return absCandidate, nil
		}
		if allowBasenameFallback && isBasename(dir) {
			return "", fileCollisionGuidedError(wrkHome, dir, absCandidate, hint)
		}
		return absCandidate, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat dir: %w", err)
	}

	if allowBasenameFallback && isBasename(dir) {
		resolved, fallbackErr := resolveBasenameFromProjects(wrkHome, dir)
		if fallbackErr != nil {
			return "", fallbackErr
		}
		if resolved != "" {
			if _, err := os.Stat(resolved); err != nil {
				if os.IsNotExist(err) {
					return "", fmt.Errorf("wrk: %s does not exist", resolved)
				}
				return "", fmt.Errorf("stat dir: %w", err)
			}
			return resolved, nil
		}
	}

	return "", fmt.Errorf("wrk: %s does not exist", absCandidate)
}

func fileCollisionGuidedError(wrkHome, basename, filePath string, hint *DirHintOptions) error {
	matches, err := storage.FindProjectsByBasename(wrkHome, basename)
	if err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "wrk: %s exists and is a file", filePath)
	if len(matches) == 0 {
		return errors.New(b.String())
	}

	fmt.Fprintf(&b, "\n%q matches registered project(s):", basename)
	for _, p := range matches {
		fmt.Fprintf(&b, "\n  %s", p)
	}

	resolvedDir := matches[0]
	if len(matches) > 1 {
		resolvedDir = "<full-path>"
	}
	fmt.Fprintf(&b, "\nuse `%s` instead", reconstructHintCommand(hint, resolvedDir))
	return errors.New(b.String())
}

func reconstructHintCommand(hint *DirHintOptions, resolvedDir string) string {
	if hint == nil {
		return "wrk " + resolvedDir
	}
	if hint.DepMode {
		return reconstructDepHint(hint.RawArgs, resolvedDir)
	}
	return reconstructInvocationHint(hint.RawArgs, hint.Positionals, resolvedDir, 0)
}

func reconstructDepHint(rawArgs []string, resolvedDir string) string {
	return reconstructInvocationHint(rawArgs, nil, resolvedDir, 1)
}

// reconstructInvocationHint rebuilds a suggested wrk command from raw CLI args,
// replacing the first positional (skipPositional=0) or --bring value (skipPositional=1).
func reconstructInvocationHint(rawArgs, positionals []string, resolvedDir string, replaceMode int) string {
	var parts []string
	parts = append(parts, "wrk")
	if replaceMode == 0 {
		parts = append(parts, resolvedDir)
	}

	pos := 0
	skipValue := false
	for _, arg := range rawArgs {
		if skipValue {
			if replaceMode == 1 {
				parts = append(parts, resolvedDir)
			} else {
				parts = append(parts, quoteHintArg(arg))
			}
			skipValue = false
			continue
		}
		if replaceMode == 0 && pos < len(positionals) && arg == positionals[pos] {
			pos++
			continue
		}
		parts = append(parts, arg)
		if arg == "--bring" {
			skipValue = true
		} else if _, ok := flagValueArgs[arg]; ok {
			skipValue = true
		}
	}
	return strings.Join(parts, " ")
}

func quoteHintArg(arg string) string {
	if strings.ContainsAny(arg, " \t") {
		return "'" + strings.ReplaceAll(arg, "'", `'"'"'`) + "'"
	}
	return arg
}

// resolveSourceWorkDir resolves the effective workDir from an optional sourceDir
// positional. When sourceDir is absent, returns the process cwd.
func resolveSourceWorkDir(origWd, sourceDir string, allowBasenameFallback bool, wrkHome string, hint *DirHintOptions) (string, error) {
	if sourceDir == "" {
		wd, err := processCwd()
		if err != nil {
			return "", fmt.Errorf("get cwd: %w", err)
		}
		return wd, nil
	}

	_ = origWd // resolveDirArg already resolves relative paths against process cwd.
	return resolveDirArg(sourceDir, allowBasenameFallback, wrkHome, hint)
}

func resolveBasenameFromProjects(wrkHome, basename string) (string, error) {
	matches, err := storage.FindProjectsByBasename(wrkHome, basename)
	if err != nil {
		return "", err
	}
	switch len(matches) {
	case 0:
		return "", nil
	case 1:
		return matches[0], nil
	default:
		return pickAmbiguousBasename(basename, matches)
	}
}

func pickAmbiguousBasename(basename string, matches []string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Multiple projects match %q:\n", basename)
	for i, p := range matches {
		fmt.Fprintf(&b, "  %d) %s\n", i+1, p)
	}
	listing := b.String()

	bypass := os.Getenv("WRK_BASENAME_CONFIRM") == "1"
	if bypass || term.IsTerminal(int(os.Stdin.Fd())) {
		n := len(matches)
		fmt.Fprint(os.Stderr, listing)
		fmt.Fprintf(os.Stderr, "Select [1-%d]: ", n)
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("wrk: read selection: %w", err)
		}
		choice, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || choice < 1 || choice > n {
			return "", fmt.Errorf("wrk: invalid selection")
		}
		return matches[choice-1], nil
	}

	return "", errors.New(strings.TrimRight(listing, "\n"))
}
