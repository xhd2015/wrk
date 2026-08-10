package wrkcli

import (
	"fmt"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
)

// runDepReplace implements wrk --dep-replace <dir>… [--dry-run].
// Edits the nearest consumer go.mod walking up from workDir (no Chdir, no tidy).
// Multi-arg is fail-fast: first error stops; prior successful writes may remain.
func runDepReplace(workDir string, paths []string, dryRun bool) error {
	if len(paths) == 0 {
		return fmt.Errorf("wrk: --dep-replace requires a directory")
	}
	consumerDir, err := findModuleRootWalking(workDir)
	if err != nil {
		return fmt.Errorf("wrk: %w", err)
	}
	for _, p := range paths {
		absDep, err := absAgainstProcessCwd(p)
		if err != nil {
			return fmt.Errorf("wrk: resolve %s: %w", p, err)
		}
		if dryRun {
			modPath, absDir, err := resolveDepModuleForReplace(absDep)
			if err != nil {
				return fmt.Errorf("wrk: %w", err)
			}
			fmt.Printf("would: dep-replace %s => %s\n", modPath, absDir)
			continue
		}
		absDir, modulePath, err := replace.ReplaceIn(consumerDir, absDep)
		if err != nil {
			return fmt.Errorf("wrk: %w", err)
		}
		fmt.Printf("dep-replace %s => %s\n", modulePath, absDir)
	}
	return nil
}

// runDepUpdate implements wrk --dep-update <dir>… [--dry-run].
// Drops replace and sets require to the latest matching tag via update.Pin.
// No tidy. Multi-arg fail-fast.
func runDepUpdate(workDir string, paths []string, dryRun bool) error {
	if len(paths) == 0 {
		return fmt.Errorf("wrk: --dep-update requires a directory")
	}
	consumerDir, err := findModuleRootWalking(workDir)
	if err != nil {
		return fmt.Errorf("wrk: %w", err)
	}
	for _, p := range paths {
		absDep, err := absAgainstProcessCwd(p)
		if err != nil {
			return fmt.Errorf("wrk: resolve %s: %w", p, err)
		}
		result, err := update.Pin(update.PinOptions{
			ConsumerDir: consumerDir,
			DepDir:      absDep,
			DryRun:      dryRun,
		})
		if err != nil {
			return fmt.Errorf("wrk: %w", err)
		}
		if dryRun {
			fmt.Printf("would: dep-update %s -> %s\n", result.ModulePath, result.Version)
			continue
		}
		if result.Tag != "" {
			fmt.Printf("dep-update %s -> %s  (tag %s)\n", result.ModulePath, result.Version, result.Tag)
		} else {
			fmt.Printf("dep-update %s -> %s\n", result.ModulePath, result.Version)
		}
	}
	return nil
}

// resolveDepModuleForReplace resolves module path + absolute dep dir without writing.
// Mirrors replace.ReplaceIn validation (exists + go module).
func resolveDepModuleForReplace(dir string) (modulePath, absDir string, err error) {
	if dir == "" {
		return "", "", fmt.Errorf("requires dir")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("no such dir: %s", dir)
	}
	absDir, err = absAgainstProcessCwd(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %v", err)
	}
	modInfo, err := resolve.GetModuleInfo(absDir)
	if err != nil {
		return "", "", err
	}
	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", absDir)
	}
	return modInfo.Module.Path, absDir, nil
}
