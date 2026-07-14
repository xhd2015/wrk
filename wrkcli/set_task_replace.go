package wrkcli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

// rewriteConsumerReplacePaths updates absolute filesystem replace targets in all
// Go modules under newRoot that still point at paths under oldRoot (the consumer
// checkout before git worktree move).
func rewriteConsumerReplacePaths(oldRoot, newRoot string) error {
	oldRoot = filepath.Clean(oldRoot)
	newRoot = filepath.Clean(newRoot)

	modules, err := scan.Scan(newRoot, scan.Options{})
	if err != nil {
		return fmt.Errorf("scan modules under %s: %w", newRoot, err)
	}

	for _, m := range modules {
		modDir := newRoot
		if m.Dir != "." {
			modDir = filepath.Join(newRoot, filepath.FromSlash(m.Dir))
		}

		for _, repl := range m.LocalFilesystemReplaces() {
			if repl.NewVersion != "" {
				continue
			}
			newTarget, ok := rewriteReplacePathIfUnder(oldRoot, newRoot, repl.NewPath)
			if !ok {
				continue
			}
			opts := &commands.GoModEditOptions{Dir: modDir, Stderr: false, Stdout: false}
			if err := commands.GoModEditReplace(repl.OldPath, newTarget, opts); err != nil {
				return fmt.Errorf("rewrite replace %s in %s: %w", repl.OldPath, modDir, err)
			}
		}
	}
	return nil
}

// rewriteReplacePathIfUnder returns the path under newRoot corresponding to
// replacePath when replacePath is an absolute path contained in oldRoot.
func rewriteReplacePathIfUnder(oldRoot, newRoot, replacePath string) (string, bool) {
	if replacePath == "" {
		return "", false
	}
	if strings.HasPrefix(replacePath, "./") || strings.HasPrefix(replacePath, "../") {
		return "", false
	}
	if !filepath.IsAbs(replacePath) {
		return "", false
	}

	cleanPath := filepath.Clean(replacePath)
	rel, err := filepath.Rel(oldRoot, cleanPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", false
	}

	return filepath.Join(newRoot, rel), true
}