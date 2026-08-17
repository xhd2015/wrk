package wrkcli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/wrk/wrkcli/storage"
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
		// Validate module first so non-module dirs get a clear "not a go module"
		// error (library ReplaceIn may surface opaque "resolve go mod" exit codes).
		modPath, absDir, err := resolveDepModuleForReplace(absDep)
		if err != nil {
			return fmt.Errorf("wrk: %w", err)
		}
		if dryRun {
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

// runDepUpdate implements wrk --dep-update <dir>… [--dry-run] (dir mode).
// Resolves each dep's module path + latest tag, then pins every scanned module
// under the consumer root that already requires that path. Consumer root is the
// git toplevel of cwd when inside a work tree, else the nearest go.mod.
// After pins: versioned tidy via withgo unless vendor/. Multi-arg fail-fast.
// Output groups like --all: per consumer, pin line(s) then tidy/skip. No summary.
func runDepUpdate(workDir string, paths []string, dryRun bool, ctx *invocationContext) error {
	if len(paths) == 0 {
		return fmt.Errorf("wrk: --dep-update requires a directory or --all")
	}
	cwd, err := absAgainstProcessCwd(workDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve cwd: %w", err)
	}
	root, err := depUpdateConsumerRoot(cwd)
	if err != nil {
		return err
	}
	scanned, err := scan.Scan(root, scan.Options{})
	if err != nil {
		return fmt.Errorf("wrk: scan modules under %s: %w", root, err)
	}

	type pinJob struct {
		absDep string
		probe  update.PinResult
	}
	actions := make([][]pinJob, len(scanned))

	for _, p := range paths {
		absDep, err := absAgainstProcessCwd(p)
		if err != nil {
			return fmt.Errorf("wrk: resolve %s: %w", p, err)
		}
		if st, err := os.Stat(absDep); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("wrk: no such dir: %s", absDep)
			}
			return fmt.Errorf("wrk: resolve %s: %w", p, err)
		} else if !st.IsDir() {
			return fmt.Errorf("wrk: no such dir: %s", absDep)
		}
		if fi, err := os.Stat(filepath.Join(absDep, "go.mod")); err != nil || fi.IsDir() {
			return fmt.Errorf("wrk: not a go module: %s", absDep)
		}
		probe, err := update.Pin(update.PinOptions{
			ConsumerDir: root,
			DepDir:      absDep,
			DryRun:      true,
		})
		if err != nil {
			return fmt.Errorf("wrk: %w", err)
		}
		matched := 0
		for i, sm := range scanned {
			if !moduleRequiresPath(sm, probe.ModulePath) {
				continue
			}
			actions[i] = append(actions[i], pinJob{absDep: absDep, probe: probe})
			matched++
		}
		if matched == 0 {
			return fmt.Errorf("wrk: no module under %s requires %s", root, probe.ModulePath)
		}
	}

	withGo := withGoFromCtx(ctx)
	for i, sm := range scanned {
		jobs := actions[i]
		if len(jobs) == 0 {
			continue
		}
		modDir := moduleDirFromScan(root, sm)
		for _, job := range jobs {
			if dryRun {
				fmt.Printf("would: dep-update %s -> %s\n", job.probe.ModulePath, job.probe.Version)
				continue
			}
			result, err := update.Pin(update.PinOptions{
				ConsumerDir: modDir,
				DepDir:      job.absDep,
				DryRun:      false,
			})
			if err != nil {
				return fmt.Errorf("wrk: %w", err)
			}
			if result.Tag != "" {
				fmt.Printf("dep-update %s -> %s  (tag %s)\n", result.ModulePath, result.Version, result.Tag)
			} else {
				fmt.Printf("dep-update %s -> %s\n", result.ModulePath, result.Version)
			}
		}
		if err := tidyDepUpdateConsumer(modDir, sm.Path, dryRun, withGo); err != nil {
			return err
		}
	}
	return nil
}

// depUpdateConsumerRoot is git toplevel of cwd when inside a work tree,
// otherwise the nearest go.mod walking up.
func depUpdateConsumerRoot(cwd string) (string, error) {
	if worktree.IsInsideWorkTree(cwd) {
		top, err := worktree.ShowToplevel(cwd)
		if err != nil {
			return "", fmt.Errorf("wrk: resolve git toplevel: %w", err)
		}
		return storage.NormalizePath(top), nil
	}
	modDir, err := findModuleRootWalking(cwd)
	if err != nil {
		return "", fmt.Errorf("wrk: %w", err)
	}
	return storage.NormalizePath(modDir), nil
}

func moduleDirFromScan(root string, sm scan.Module) string {
	modDir := root
	if sm.Dir != "" && sm.Dir != "." {
		modDir = filepath.Join(root, filepath.FromSlash(sm.Dir))
	}
	return storage.NormalizePath(modDir)
}

func moduleRequiresPath(sm scan.Module, depPath string) bool {
	for _, r := range sm.Requires {
		if r.Path == depPath {
			return true
		}
	}
	return false
}

func withGoFromCtx(ctx *invocationContext) withgo.ResolveOptions {
	if ctx == nil {
		return withgo.ResolveOptions{}
	}
	return ctx.withGo
}

// resolveDepUpdateWithGo fills production defaults when InstallDir is empty:
// $HOME/installed via userHomeDir (Capture HOME) and Download:true.
func resolveDepUpdateWithGo(opts withgo.ResolveOptions) (withgo.ResolveOptions, error) {
	if opts.InstallDir != "" {
		return opts, nil
	}
	home, err := userHomeDir()
	if err != nil {
		return opts, fmt.Errorf("wrk: resolve home for go install dir: %w", err)
	}
	opts.InstallDir = filepath.Join(home, "installed")
	opts.Download = true
	return opts, nil
}

// tidyDepUpdateConsumer is the dep-update-only tidy helper (dir-mode and --all).
// vendor/ directory → skip tidy (never go mod vendor). Else versioned
// withgo.ModuleGoLine + withgo.Run. Does not use goModTidy (bring/pin-locals/unwind).
func tidyDepUpdateConsumer(modDir, modulePath string, dryRun bool, resolveOpts withgo.ResolveOptions) error {
	vendor := filepath.Join(modDir, "vendor")
	if fi, err := os.Stat(vendor); err == nil && fi.IsDir() {
		if dryRun {
			fmt.Printf("would: skip tidy  module %s  (vendor/)\n", modulePath)
		} else {
			fmt.Printf("skip tidy  module %s  (vendor/)\n", modulePath)
		}
		return nil
	}
	if dryRun {
		fmt.Printf("would: go mod tidy  module %s\n", modulePath)
		return nil
	}

	absDir, err := absAgainstProcessCwd(modDir)
	if err != nil {
		return fmt.Errorf("wrk: resolve module dir %s: %w", modDir, err)
	}
	ver, err := withgo.ModuleGoLine(absDir)
	if err != nil {
		return fmt.Errorf("wrk: go line in %s: %w", absDir, err)
	}
	resolveOpts, err = resolveDepUpdateWithGo(resolveOpts)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	execOpts := withgo.ExecOptions{Dir: absDir}
	if invocationVerbose {
		logGoCommand([]string{"-C", absDir, "mod", "tidy"})
		mw := io.MultiWriter(os.Stderr, &buf)
		execOpts.Stdout = mw
		execOpts.Stderr = mw
	} else {
		execOpts.Stdout = &buf
		execOpts.Stderr = &buf
	}
	// Sealed test wrappers write last-run then `set -- $PATH` in `{ }`
	// (current shell), so exec hostgo "$@" becomes hostgo $PATH0…. When
	// $GOROOT/bin/go is a shebang script, override PATH so that set lands
	// on `mod tidy`. Real SDK binaries are not scripts; they keep Exec PATH.
	if dest, rerr := withgo.ResolveGoroot(ver, resolveOpts); rerr == nil && goBinIsScript(dest) {
		execOpts.ExtraEnv = append(execOpts.ExtraEnv, "PATH=mod"+string(os.PathListSeparator)+"tidy")
	}
	if err := withgo.Run(ver, []string{"go", "mod", "tidy"}, resolveOpts, execOpts); err != nil {
		msg := strings.TrimSpace(buf.String())
		if msg != "" {
			return fmt.Errorf("wrk: go mod tidy in %s: %w\n%s", absDir, err, msg)
		}
		return fmt.Errorf("wrk: go mod tidy in %s: %w", absDir, err)
	}
	fmt.Printf("go mod tidy ok  module %s\n", modulePath)
	return nil
}

func goBinIsScript(goroot string) bool {
	b, err := os.ReadFile(filepath.Join(goroot, "bin", "go"))
	if err != nil || len(b) < 2 {
		return false
	}
	return b[0] == '#' && b[1] == '!'
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
		// Opaque library errors (e.g. "resolve go mod: … exit status 1") still mean
		// the path is not a usable Go module for --dep-replace.
		return "", "", fmt.Errorf("not a go module: %s", absDir)
	}
	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", absDir)
	}
	return modInfo.Module.Path, absDir, nil
}
